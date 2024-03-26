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
	ServicePort string `envconfig:"SERVICE_PORT"`
	ServiceCrt  string `envconfig:"SERVICE_CRT"`
	ServiceKey  string `envconfig:"SERVICE_KEY"`
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
		Addr:    cfg.ServicePort,
		Handler: g,
	}

	err = tlsConfig.ListenAndServeTLS(cfg.ServiceCrt, cfg.ServiceKey)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
