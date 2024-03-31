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
	ServicePort  string `envconfig:"SERVICE_PORT"`
	ServiceCrt   string `envconfig:"SERVICE_CRT"`
	ServiceKey   string `envconfig:"SERVICE_KEY"`
	ApiBaseUrl   string `envconfig:"API_BASE_URL"`
	ApiKey       string `envconfig:"API_KEY"`
	DockerHost   string `envconfig:"DOCKER_HOST"`
	DhUsername   string `envconfig:"DH_USERNAME"`
	DhPassword   string `envconfig:"DH_PASSWORD"`
	RepoUsername string `envconfig:"REPO_USERNAME"`
	RepoPassword string `envconfig:"REPO_PASSWORD"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	if len(os.Args) > 1 {
		arg1 := os.Args[1]

		switch arg1 {
		case "start":
			initService(cfg)
		case "version":
			fmt.Println(Version)
		case "env":
			fmt.Printf("%+v\n", cfg)
		default:
			fmt.Sprintln("Uknown command line arguments", os.Args)
		}
	} else {
		fmt.Println("Please provide a command line argument")
	}

}

func initService(cfg Config) {
	g := gin.Default()

	g.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // Maximum cache age (12 hours)
	}))

	a := api.NewScheduler(api.SchedulerConfig{
		ApiBaseUrl:   cfg.ApiBaseUrl,
		ApiKey:       cfg.ApiKey,
		DockerHost:   cfg.DockerHost,
		DhUsername:   cfg.DhUsername,
		DhPassword:   cfg.DhPassword,
		RepoUsername: cfg.RepoUsername,
		RepoPassword: cfg.RepoPassword,
	})

	// Define API routes
	g.POST("/streamWebhook", a.StreamWebhookHandler())
	g.GET("/healthCheck", a.HealthCheckHandler())
	g.GET("/serviceLogs", a.GetServiceLog())
	g.GET("/diskInfo", a.GetDiskInfo())
	g.DELETE("/images", a.DeleteImages())
	g.DELETE("/builderCache", a.DeleteBuilderCache())
	g.GET("/network/:name", a.GetNetwork())
	g.GET("/networks", a.GetNetworks())
	g.DELETE("/network/:name", a.DeleteNetwork())
	g.POST("/network", a.CreateNetwork())
	g.GET("/service/:name", a.GetService())
	g.GET("/services", a.GetServices())
	g.DELETE("/service/:name", a.DeleteService())
	g.PUT("/service", a.UpdateService())

	// OPTIONS routes for CORS preflight requests
	g.OPTIONS("/streamWebhook", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/healthCheck", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/serviceLogs", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/diskInfo", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/images", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/builderCache", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/networks", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/network", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/network/:name", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/service/:name", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/services", func(c *gin.Context) { c.Status(http.StatusOK) })
	g.OPTIONS("/service", func(c *gin.Context) { c.Status(http.StatusOK) })

	tlsConfig := &http.Server{
		Addr:    cfg.ServicePort,
		Handler: g,
	}

	err := tlsConfig.ListenAndServeTLS(cfg.ServiceCrt, cfg.ServiceKey)
	if err != nil {
		fmt.Println("Error starting service:", err)
	}
}
