package task

import (
	"encoding/json"

	"github.com/kelseyhightower/envconfig"
	"github.com/more-than-code/deploybot-service-api/model"
	"github.com/more-than-code/deploybot-service-launcher/util"
)

type RunnerConfig struct {
	ProjectsPath string `envconfig:"PROJECTS_PATH"`
	DockerHost   string `envconfig:"DOCKER_HOST"`
}

type Runner struct {
	cfg     RunnerConfig
	cHelper *util.ContainerHelper
}

func NewRunner() *Runner {
	var cfg RunnerConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	return &Runner{cfg: cfg, cHelper: util.NewContainerHelper(cfg.DockerHost)}
}

func (r *Runner) DoTask(t model.Task, arguments []string) error {

	var c model.DeployConfig

	bs, err := json.Marshal(t.Config)

	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, &c)

	if err != nil {
		return err
	}

	go func() {
		r.cHelper.StartContainer(&c)
	}()

	return nil
}
