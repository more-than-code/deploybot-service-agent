package main

import (
	"fmt"
	"net/http"
	"os"

	"deploybot-service-agent/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
)

var Version string // This will be set during build using -ldflags

type Config struct {
	ServicePort string `envconfig:"SERVICE_PORT"`
	ServiceCrt  string `envconfig:"SERVICE_CRT"`
	ServiceKey  string `envconfig:"SERVICE_KEY"`
}

func main() {
	if len(os.Args) > 1 {
		arg1 := os.Args[1]

		fmt.Println("arg1:", arg1)

		switch arg1 {
		case "start":
			initService()
		case "version":
			fmt.Println(Version)
		default:
			fmt.Println("Uknown command line arguments", os.Args)
		}
	} else {
		fmt.Println("Please provide a command line argument")
	}

}

func initService() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	g := gin.Default()

	g.Use(cors.Default())

	a := api.NewScheduler()
	g.POST("/streamWebhook", a.StreamWebhookHandler())
	g.GET("/healthCheck", a.HealthCheckHandler())

	g.GET("/serviceLogs", a.GetServiceLog())
	g.GET("/diskInfo", a.GetDiskInfo())

	g.OPTIONS("/images", func(ctx *gin.Context) {})
	g.DELETE("/images", a.DeleteImages())

	g.GET("/network/:name", a.GetNetwork())
	g.GET("/networks", a.GetNetworks())
	g.OPTIONS("/network", func(ctx *gin.Context) {})
	g.DELETE("/network/:name", a.DeleteNetwork())
	g.POST("/network", a.PostNetwork())

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
