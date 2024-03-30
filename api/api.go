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

func (s *Scheduler) PostNetwork() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input model.CreateNetworkInput
		err := ctx.BindJSON(&input)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.CreateNetworkResponse{Msg: err.Error(), Code: types.CodeClientError})
		}

		name := input.Name

		networkId, err := s.cHelper.CreateNetwork(ctx, name)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.GetNetworkResponse{Msg: err.Error(), Code: types.CodeServerError})
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
			ctx.JSON(http.StatusBadRequest, model.GetNetworkResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.GetNetworkResponse{Payload: &model.Network{Name: name, Id: networkId}})
	}
}

func (s *Scheduler) GetNetworks() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		networks, err := s.cHelper.GetNetworks(ctx)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, model.GetNetworksResponse{Msg: err.Error(), Code: types.CodeServerError})
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
			ctx.JSON(http.StatusBadRequest, model.DeleteNetworkResponse{Msg: err.Error(), Code: types.CodeServerError})
			return
		}
		ctx.JSON(http.StatusOK, model.DeleteNetworkResponse{})
	}
}

func (s *Scheduler) HealthCheckHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
