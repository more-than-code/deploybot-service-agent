package main

import (
	"fmt"
	"net/http"

	"deploybot-service-launcher/task"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerPort string `envconfig:"SERVER_PORT"`
	ServerCrt  string `envconfig:"SERVER_CRT"`
	ServerKey  string `envconfig:"SERVER_KEY"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	g := gin.Default()

	g.Use(cors.Default())

	t := task.NewScheduler()
	g.POST("/streamWebhook", t.StreamWebhookHandler())
	g.GET("/healthCheck", t.HealthCheckHandler())

	tlsConfig := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: g,
	}

	err = tlsConfig.ListenAndServeTLS(cfg.ServerCrt, cfg.ServerKey)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
