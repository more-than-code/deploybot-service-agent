package task

import (
	"testing"

	types "deploybot-service-launcher/deploybot-types"
)

func TestRunContainer(t *testing.T) {
	env := []string{}

	r := NewRunner()
	err := r.DoTask(types.Task{Type: types.TaskDeploy, Config: types.DeployConfig{ExposedPort: "9000", HostPort: "9000", Env: env, ImageName: "binartist/mo-service-graph", ImageTag: "latest", ServiceName: "graph", AutoRemove: false}}, nil)

	if err != nil {
		t.Error(err)
	}

}
