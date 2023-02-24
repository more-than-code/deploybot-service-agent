package task

import (
	"testing"

	"github.com/more-than-code/deploybot-service-api/model"
)

func TestRunContainer(t *testing.T) {
	env := []string{}

	r := NewRunner()
	err := r.DoTask(model.Task{Type: model.TaskDeploy, Config: model.DeployConfig{ExposedPort: "9000", HostPort: "9000", Env: env, ImageName: "binartist/mo-service-graph", ImageTag: "latest", ServiceName: "graph", AutoRemove: false}}, nil)

	if err != nil {
		t.Error(err)
	}

}
