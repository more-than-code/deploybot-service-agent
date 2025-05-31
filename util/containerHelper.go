package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"deploybot-service-agent/model"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DhCredentials struct {
	Username string
	Password string
}

type ContainerHelper struct {
	cli  *client.Client
	cred DhCredentials
}

type ChLogsOptions struct {
	ShowStdout bool
	ShowStderr bool
	Follow     bool
	Details    bool
	Timestamps bool
	Tail       string
	Until      string
	Since      string
}

func NewContainerHelper(dockerHost string, cred DhCredentials) *ContainerHelper {
	cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &ContainerHelper{cli, cred}
}

func (h *ContainerHelper) StartContainer(cfg *model.DeployConfig) {
	ctx := context.Background()

	h.cli.ContainerStop(ctx, cfg.ServiceName, container.StopOptions{})
	h.cli.ContainerRemove(ctx, cfg.ServiceName, container.RemoveOptions{})

	imageNameTag := cfg.ImageName + ":" + cfg.ImageTag
	reader, err := h.cli.ImagePull(ctx, imageNameTag, image.PullOptions{})
	if err != nil {
		log.Print(err)
		return
	}
	io.Copy(os.Stdout, reader)

	cConfig := &container.Config{
		Image: imageNameTag,
		Env:   cfg.Env,
	}

	if cfg.Command != "" {
		cConfig.Cmd = strings.Split(cfg.Command, " ")
	}

	if cfg.RestartPolicy.Name == "" {
		cfg.RestartPolicy.Name = "on-failure"
	}

	hConfig := &container.HostConfig{
		AutoRemove:    cfg.AutoRemove,
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyMode(cfg.RestartPolicy.Name), MaximumRetryCount: cfg.RestartPolicy.MaximumRetryCount},
		Links:         cfg.Links,
	}

	if cfg.Ports != nil {
		cConfig.ExposedPorts = nat.PortSet{}
		for e := range cfg.Ports {
			cConfig.ExposedPorts[nat.Port(e+"/tcp")] = struct{}{}
		}

		hConfig.PortBindings = nat.PortMap{}
		for e, h := range cfg.Ports {
			hConfig.PortBindings[nat.Port(e+"/tcp")] = []nat.PortBinding{{HostPort: h, HostIP: ""}}
		}
	}

	if cfg.VolumeMounts != nil {
		for s, t := range cfg.VolumeMounts {
			hConfig.Mounts = append(hConfig.Mounts, mount.Mount{Type: mount.TypeBind, Source: s, Target: t})
		}
	}

	nConfig := &network.NetworkingConfig{}

	if cfg.Networks != nil {
		for n, i := range cfg.Networks {
			nConfig.EndpointsConfig = map[string]*network.EndpointSettings{n: {NetworkID: i}}
		}
	}

	resp, err := h.cli.ContainerCreate(ctx, cConfig, hConfig, nConfig, nil, cfg.ServiceName)
	if err != nil {
		log.Print(err)
		return
	}

	if err := h.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Print(err)
		return
	}
}

func (h *ContainerHelper) RestartContainer(ctx context.Context, containerName string) error {
	return h.cli.ContainerRestart(ctx, containerName, container.StopOptions{})
}

func (h *ContainerHelper) LogContainer(ctx context.Context, containerName string, logsOptions ChLogsOptions) (io.ReadCloser, error) {
	clogsOptions := container.LogsOptions{
		ShowStdout: logsOptions.ShowStdout,
		ShowStderr: logsOptions.ShowStderr,
		Follow:     logsOptions.Follow,
		Details:    logsOptions.Details,
		Timestamps: logsOptions.Timestamps,
		Tail:       logsOptions.Tail,
		Until:      logsOptions.Until,
		Since:      logsOptions.Since,
	}

	return h.cli.ContainerLogs(ctx, containerName, clogsOptions)
}

func (h *ContainerHelper) RemoveContainer(ctx context.Context, containerName string) error {
	return h.cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})
}

func (h *ContainerHelper) StopContainer(ctx context.Context, containerName string) error {
	return h.cli.ContainerStop(ctx, containerName, container.StopOptions{})
}

func (h *ContainerHelper) GetContainer(ctx context.Context, containerId string) (interface{}, error) {
	return h.cli.ContainerInspect(ctx, containerId)
}

func (h *ContainerHelper) GetContainers(ctx context.Context) ([]interface{}, error) {
	containers, err := h.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	var res []interface{}
	for _, c := range containers {
		res = append(res, c)
	}

	return res, nil
}

func (h *ContainerHelper) CreateNetwork(ctx context.Context, networkName string) (string, error) {
	res, err := h.cli.NetworkCreate(ctx, networkName, types.NetworkCreate{Driver: "bridge"})

	if err != nil {
		return "", err
	}

	return res.ID, nil
}

func (h *ContainerHelper) GetNetworkId(ctx context.Context, networkName string) (string, error) {
	res, err := h.cli.NetworkInspect(ctx, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return "", err
	}

	return res.ID, nil
}

func (h *ContainerHelper) GetNetworks(ctx context.Context) ([]model.Network, error) {
	networks, err := h.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	var res []model.Network
	for _, n := range networks {
		res = append(res, model.Network{Name: n.Name, Id: n.ID})
	}

	return res, nil
}

func (h *ContainerHelper) RemoveNetwork(ctx context.Context, networkName string) error {
	return h.cli.NetworkRemove(ctx, networkName)
}

func (h *ContainerHelper) RemoveImages(ctx context.Context) error {
	images, err := h.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return err
	}

	for _, img := range images {
		items, err := h.cli.ImageRemove(ctx, img.ID, image.RemoveOptions{Force: true, PruneChildren: true})

		if err != nil {
			return err
		}

		log.Println(items)
	}

	return nil
}

func (h *ContainerHelper) RemoveBuilderCache(ctx context.Context) error {
	report, err := h.cli.BuildCachePrune(ctx, types.BuildCachePruneOptions{})
	if err != nil {
		return err
	}

	log.Println(report)

	return nil
}

func (h *ContainerHelper) BuildImage(buildContext io.Reader, buidOptions *types.ImageBuildOptions) error {
	buildResponse, err := h.cli.ImageBuild(context.Background(), buildContext, *buidOptions)

	if err != nil {
		return err
	}

	defer buildResponse.Body.Close()

	io.Copy(os.Stdout, buildResponse.Body)

	return nil
}

func (h *ContainerHelper) PushImage(name string) error {
	authConfig := registry.AuthConfig{
		Username: h.cred.Username,
		Password: h.cred.Password,
	}
	encodedJSON, _ := json.Marshal(authConfig)
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	res, err := h.cli.ImagePush(context.Background(), name, image.PushOptions{RegistryAuth: authStr})

	if err != nil {
		return err
	}

	defer res.Close()
	io.Copy(os.Stdout, res)
	return nil
}
