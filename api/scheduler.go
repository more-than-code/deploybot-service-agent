package api

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	types "deploybot-service-agent/deploybot-types"
	"deploybot-service-agent/model"
	"deploybot-service-agent/util"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

var gTicker *time.Ticker
var gEventQueue = list.New()

type SchedulerConfig struct {
	ApiBaseUrl string `envconfig:"API_BASE_URL"`
	ApiKey     string `envconfig:"API_KEY"`
	DockerHost string `envconfig:"DOCKER_HOST"`
}

type Scheduler struct {
	cHelper *util.ContainerHelper
	cfg     SchedulerConfig
}

func NewScheduler() *Scheduler {
	var cfg SchedulerConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	return &Scheduler{cHelper: util.NewContainerHelper(cfg.DockerHost), cfg: cfg}
}

func (s *Scheduler) PushEvent(e types.Event) {
	gEventQueue.PushBack(e)
}

func (s *Scheduler) PullEvent() types.Event {
	e := gEventQueue.Front()

	gEventQueue.Remove(e)

	return e.Value.(types.Event)
}

func (s *Scheduler) updateTaskStatus(pipelineId, taskId types.ObjectId, status string) {
	body, _ := json.Marshal(types.UpdateTaskStatusInput{
		PipelineId: pipelineId,
		TaskId:     taskId,
		Task:       struct{ Status string }{Status: status}})

	req, _ := http.NewRequest("PUT", s.cfg.ApiBaseUrl+"/taskStatus", bytes.NewReader(body))
	req.Header.Set("X-Api-Key", s.cfg.ApiKey)
	http.DefaultClient.Do(req)
}

func (s *Scheduler) ProcessPostTask(pipelineId, taskId types.ObjectId, status string) {
	body, _ := json.Marshal(types.UpdateTaskStatusInput{
		PipelineId: pipelineId,
		TaskId:     taskId,
		Task:       struct{ Status string }{Status: status}})

	req, _ := http.NewRequest("PUT", s.cfg.ApiBaseUrl+"/taskStatus", bytes.NewReader(body))
	req.Header.Set("X-Api-Key", s.cfg.ApiKey)
	http.DefaultClient.Do(req)
}

func (s *Scheduler) StreamWebhookHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body, _ := io.ReadAll(ctx.Request.Body)

		var sw types.StreamWebhook
		json.Unmarshal(body, &sw)

		log.Println(sw.Payload)

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/task?pid=%s&id=%s", s.cfg.ApiBaseUrl, sw.Payload.PipelineId.Hex(), sw.Payload.TaskId.Hex()), nil)
		req.Header.Set("X-Api-Key", s.cfg.ApiKey)

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusBadRequest, types.WebhookResponse{Msg: err.Error(), Code: types.CodeServerError})

			return
		}

		if res.StatusCode != 200 {
			log.Println(res.Body)
			ctx.JSON(http.StatusBadRequest, types.WebhookResponse{Msg: types.MsgClientError, Code: types.CodeClientError})

			return
		}

		body, _ = io.ReadAll(res.Body)

		var tRes types.GetTaskResponse
		json.Unmarshal(body, &tRes)

		task := tRes.Payload.Task

		var timer *time.Timer
		if task.Timeout > 0 {
			timer = s.cleanUp(time.Minute*time.Duration(task.Timeout), func() {
				s.updateTaskStatus(sw.Payload.PipelineId, task.Id, types.TaskTimedOut)
			})
		}

		s.updateTaskStatus(sw.Payload.PipelineId, task.Id, types.TaskInProgress)
		ctx.JSON(http.StatusOK, types.WebhookResponse{})

		go func() {
			err := s.DoTask(task.Config, sw.Payload.Arguments)

			if timer != nil {
				timer.Stop()
			}

			if err != nil {
				log.Println(err)
				s.ProcessPostTask(sw.Payload.PipelineId, task.Id, types.TaskFailed)
			} else {
				s.ProcessPostTask(sw.Payload.PipelineId, task.Id, types.TaskDone)
			}
		}()
	}
}

func (s *Scheduler) DoTask(conf interface{}, arguments []string) error {

	var c model.DeployConfig

	bs, err := json.Marshal(conf)

	if err != nil {
		return err
	}

	err = json.Unmarshal(bs, &c)

	if err != nil {
		return err
	}

	if c.Files != nil {
		for name, content := range c.Files {
			err = util.WriteToFile(name, content)
			if err != nil {
				return err
			}
		}
	}

	if c.VolumeMounts != nil {
		for srcDir := range c.VolumeMounts {
			err = util.CreateDirsIfNotExist(srcDir)
			if err != nil {
				return err
			}
		}
	}

	go func() {
		s.cHelper.StartContainer(&c)
	}()

	return nil
}

func (s *Scheduler) HealthCheckHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
func (s *Scheduler) cleanUp(delay time.Duration, job func()) *time.Timer {
	t := time.NewTimer(delay)
	go func() {
		for range t.C {
			job()
		}
	}()

	return t
}
