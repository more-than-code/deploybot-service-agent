package main

import (
	"fmt"

	"deploybot-service-launcher/task"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerPort int `envconfig:"SERVER_PORT"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	g := gin.Default()

	t := task.NewScheduler()
	g.POST("/streamWebhook", t.StreamWebhookHandler())
	g.GET("/serviceLogs", t.Runner.ServiceLogHandler())
	g.GET("/healthCheck", t.HealthCheckHandler())

	g.Run(fmt.Sprintf(":%d", cfg.ServerPort))
}
