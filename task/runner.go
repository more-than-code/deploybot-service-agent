package task

import (
	"bytes"
	"encoding/json"
	"net/http"

	types "deploybot-service-launcher/deploybot-types"
	"deploybot-service-launcher/util"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"
)

type Runner struct {
	cHelper *util.ContainerHelper
}

func NewRunner() *Runner {

	return &Runner{util.NewContainerHelper()}
}

func (r *Runner) DoTask(t types.Task, arguments []string) error {

	var c types.DeployConfig

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

func (s *Runner) ServiceLogHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Query("name")
		out, err := s.cHelper.LogContainer(ctx, name)

		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		var buf bytes.Buffer

		_, err = stdcopy.StdCopy(&buf, &buf, out)

		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.String(http.StatusOK, buf.String())
	}
}
