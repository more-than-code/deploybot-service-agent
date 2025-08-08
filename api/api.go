package api

import (
	"bytes"
	types "deploybot-service-agent/deploybot-types"
	"deploybot-service-agent/model"
	"deploybot-service-agent/util"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"
)

func (s *Scheduler) GetServiceLog() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Query("name")
		if name == "" {
			ctx.String(http.StatusBadRequest, "Container name is required")
			return
		}

		ShowStdout := ctx.Query("showStdout") == "true"
		ShowStderr := ctx.Query("showStderr") == "true"
		Follow := ctx.Query("follow") == "true"
		Details := ctx.Query("details") == "true"
		Timestamps := ctx.Query("timestamps") == "true"
		Tail := ctx.Query("tail")
		until := ctx.Query("until")
		since := ctx.Query("since")

		chLogsOptions := util.ChLogsOptions{
			ShowStdout: ShowStdout,
			ShowStderr: ShowStderr,
			Follow:     Follow,
			Details:    Details,
			Timestamps: Timestamps,
			Tail:       Tail,
			Until:      until,
			Since:      since,
		}

		out, err := s.cHelper.LogContainer(ctx, name, chLogsOptions)

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

func (s *Scheduler) GetDiskInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Query("path")
		info, err := util.GetDiskInfo(path)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.GetDiskInfoResponse{Msg: err.Error(), Code: http.StatusInternalServerError})
			return
		}

		ctx.JSON(http.StatusOK, model.GetDiskInfoResponse{Payload: info})
	}
}

func (s *Scheduler) DeleteImages() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := s.cHelper.RemoveImages(ctx)

		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.String(http.StatusOK, "OK")
	}
}

func (s *Scheduler) DeleteBuilderCache() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := s.cHelper.RemoveBuilderCache(ctx)

		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.String(http.StatusOK, "OK")
	}
}

func (s *Scheduler) CreateNetwork() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input model.CreateNetworkInput
		err := ctx.BindJSON(&input)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.CreateNetworkResponse{Msg: err.Error(), Code: types.CodeServerError})
		}

		name := input.Name

		networkId, err := s.cHelper.CreateNetwork(ctx, name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.GetNetworkResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.GetNetworkResponse{Payload: &model.Network{Name: name, Id: networkId}})
	}
}

func (s *Scheduler) GetNetwork() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Param("name")

		networkId, err := s.cHelper.GetNetworkId(ctx, name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.GetNetworkResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.GetNetworkResponse{Payload: &model.Network{Name: name, Id: networkId}})
	}
}

func (s *Scheduler) GetNetworks() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		networks, err := s.cHelper.GetNetworks(ctx)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.GetNetworksResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.GetNetworksResponse{Payload: networks})
	}
}

func (s *Scheduler) DeleteNetwork() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Param("name")

		err := s.cHelper.RemoveNetwork(ctx, name)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.DeleteNetworkResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.DeleteNetworkResponse{})
	}
}

func (s *Scheduler) CreateService() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var deployConfig model.DeployConfig
		if err := ctx.ShouldBindJSON(&deployConfig); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := s.cHelper.StartContainer(&deployConfig)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "deployment started"})
	}
}

func (s *Scheduler) UpdateService() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input model.UpdateServiceInput
		err := ctx.BindJSON(&input)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.ApiResponse{Msg: err.Error(), Code: types.CodeServerError})
		}

		if input.Restarting {
			err = s.cHelper.RestartContainer(ctx, input.Name)
		} else if !input.Running {
			err = s.cHelper.StopContainer(ctx, input.Name)
		}

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, model.ApiResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}

		ctx.JSON(http.StatusOK, model.ApiResponse{})
	}
}

func (s *Scheduler) DeleteService() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Query("name")

		err := s.cHelper.RemoveContainer(ctx, name)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.ApiResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.ApiResponse{})
	}
}

func (s *Scheduler) GetService() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		name := ctx.Query("name")

		res, err := s.cHelper.GetContainer(ctx, name)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.ApiResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.ApiResponse{Payload: res})
	}
}

func (s *Scheduler) GetServices() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := s.cHelper.GetContainers(ctx)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.ApiResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.ApiResponse{Payload: res})
	}

}

func (s *Scheduler) HealthCheckHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
