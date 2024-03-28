package main

import (
	"fmt"
	"net/http"

	"deploybot-service-launcher/scheduler"

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

	t := scheduler.NewScheduler()
	g.POST("/streamWebhook", t.StreamWebhookHandler())
	g.GET("/healthCheck", t.HealthCheckHandler())

	g.GET("/network/:name", t.GetNetwork())
	g.DELETE("/network/:name", t.DeleteNetwork())
	g.POST("/network", t.PostNetwork())

	tlsConfig := &http.Server{
		Addr:    cfg.ServicePort,
		Handler: g,
	}

	fmt.Println("SERVICE_CRT:", cfg.ServiceCrt)
	fmt.Println("SERVICE_KEY:", cfg.ServiceKey)

	err = tlsConfig.ListenAndServeTLS(cfg.ServiceCrt, cfg.ServiceKey)
	if err != nil {
		fmt.Println("Error starting service:", err)
	}
}
