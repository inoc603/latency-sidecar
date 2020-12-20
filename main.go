package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("LATENCY")
	viper.AutomaticEnv()
	viper.SetDefault("port", "8080")
	viper.SetDefault("interface", "eth0")

	app := createServer(&TcLatencySetter{}, viper.GetString("interface"))
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
			c.AbortWithError(
				500,
				fmt.Errorf("failed to set latency: %w", err),
			)
			return
		}

		c.String(200, fmt.Sprintf("latency set to %s\n", d))
	})

	return app
}
