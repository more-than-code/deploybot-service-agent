package util

import (
	"context"
	"io"
	"log"
	"os"

	types "deploybot-service-launcher/deploybot-types"

	dTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kelseyhightower/envconfig"
)

type ContainerHelperConfig struct {
	DhUsername string `envconfig:"DH_USERNAME"`
	DhPassword string `envconfig:"DH_PASSWORD"`
	DockerHost string `envconfig:"DOCKER_HOST"`
}

type ContainerHelper struct {
	cli *client.Client
	cfg ContainerHelperConfig
}

func NewContainerHelper() *ContainerHelper {
	var cfg ContainerHelperConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	cli, err := client.NewClientWithOpts(client.WithHost(cfg.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &ContainerHelper{cli: cli, cfg: cfg}
}

func (h *ContainerHelper) StartContainer(cfg *types.DeployConfig) {
	ctx := context.Background()

	h.cli.ContainerStop(ctx, cfg.ServiceName, container.StopOptions{})
	h.cli.ContainerRemove(ctx, cfg.ServiceName, container.RemoveOptions{})

	imageNameTag := cfg.ImageName + ":" + cfg.ImageTag
	reader, err := h.cli.ImagePull(ctx, imageNameTag, dTypes.ImagePullOptions{})
	if err != nil {
		log.Print(err)
		return
	}
	io.Copy(os.Stdout, reader)

	cConfig := &container.Config{
		Image: imageNameTag,
		Env:   cfg.Env,
	}

	if cfg.ExposedPort != "" {
		cConfig.ExposedPorts = nat.PortSet{nat.Port(cfg.ExposedPort + "/tcp"): struct{}{}}
	}

	if cfg.RestartPolicy.Name == "" {
		cfg.RestartPolicy.Name = "on-failure"
	}

	hConfig := &container.HostConfig{
		AutoRemove:    cfg.AutoRemove,
		RestartPolicy: cfg.RestartPolicy,
	}

	if cfg.HostPort != "" {
		hConfig.PortBindings = nat.PortMap{nat.Port(cfg.ExposedPort + "/tcp"): []nat.PortBinding{{HostPort: cfg.HostPort, HostIP: "0.0.0.0"}}}
	}

	if cfg.MountSource != "" && cfg.MountTarget != "" {
		hConfig.Mounts = []mount.Mount{{Type: "bind", Source: cfg.MountSource, Target: cfg.MountTarget}}
	}

	nConfig := &network.NetworkingConfig{}

	if cfg.NetworkName != "" && cfg.NetworkId != "" {
		nConfig.EndpointsConfig = map[string]*network.EndpointSettings{cfg.NetworkName: {NetworkID: cfg.NetworkId}}
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

	// statusCh, errCh := h.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	// select {
	// case err := <-errCh:
	// 	if err != nil {
	// 		log.Print(err)
	// 		return
	// 	}
	// case <-statusCh:
	// }

	// out, err := h.cli.ContainerLogs(ctx, resp.ID, dTypes.ContainerLogsOptions{ShowStdout: true})
	// if err != nil {
	// 	log.Print(err)
	// 	return
	// }

	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func (h *ContainerHelper) RestartContainer(cfg *types.RestartConfig) error {
	return h.cli.ContainerRestart(context.Background(), cfg.ServiceName, container.StopOptions{})
}

func (h *ContainerHelper) LogContainer(ctx context.Context, containerName string) (io.ReadCloser, error) {
	return h.cli.ContainerLogs(ctx, containerName, container.LogsOptions{ShowStdout: true})
}

func (h *ContainerHelper) RemoveContainer(ctx context.Context, containerName string) error {
	return h.cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})
}

func (h *ContainerHelper) StopContainer(ctx context.Context, containerName string) error {
	return h.cli.ContainerStop(ctx, containerName, container.StopOptions{})
}
