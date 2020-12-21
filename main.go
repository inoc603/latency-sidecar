package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// for simplicity, we configure the agent with env variables
	viper.SetEnvPrefix("LATENCY")
	viper.AutomaticEnv()
	viper.SetDefault("port", "8080")
	viper.SetDefault("interface", "eth0")
	viper.SetDefault("handler", "tc")

	var setter LatencySetter
	switch viper.GetString("handler") {
	case "netlink":
		log.Println("using netlink to set latency")
		setter = NetlinkLatencySetter{}
	default:
		log.Println("using tc to set latency")
		setter = TcLatencySetter{}
	}

	app := createServer(setter, viper.GetString("interface"))
	app.Run(net.JoinHostPort("", viper.GetString("port")))
}

type LatencySetter interface {
	// SetLatency sets extra latency of the given interface
	SetLatency(iname string, latency time.Duration) error
}

func createServer(m LatencySetter, netInterface string) *gin.Engine {
	// for simplicity, we use the gin framework here.
	app := gin.New()

	app.GET("/latency/:latency", func(c *gin.Context) {
		d, err := time.ParseDuration(c.Param("latency"))
		if err != nil {
			c.AbortWithError(
				400,
				fmt.Errorf("malformed latency format %s: %w", c.Param("latency"), err),
			)
			return
		}

		if err := m.SetLatency(netInterface, d); err != nil {
			log.Printf("failed to set latency to %s: %v\n", d, err)
			c.AbortWithError(
				500,
				fmt.Errorf("failed to set latency: %w", err),
			)
			return
		}

		log.Printf("latency set to %s\n", d)

		c.String(200, fmt.Sprintf("latency set to %s\n", d))
	})

	return app
}
