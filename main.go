package main

import (
	"net/http"

	"github.com/dstockton/kafka-video-broadcast/pkg/utils"
	"github.com/dstockton/kafka-video-broadcast/pkg/video"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/olahol/melody.v1"
)

var Version = "dirty"
var BuildTime = "dirty"

func init() {
	utils.InitialiseLogging()
}

func main() {
	r, httpPort := getRouter()
	r.Run(":" + httpPort)
}

func getRouter() (*gin.Engine, string) {
	log.Infof("www-api v%v (built: %v)...", Version, BuildTime)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	m := melody.New()

	r.Static("/static", "./static")

	r.GET("/ping", func(c *gin.Context) {
		log.Debug("Received PING")
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    Version,
			"build_time": BuildTime,
		})
	})

	r.GET("/video/connections", func(c *gin.Context) {
		video.VideoConnections(
			c.Writer,
			c.Request,
			utils.GetEnv("BOOTSTRAP_SERVERS", ""),
			utils.GetEnv("SASL_MECHANISMS", "PLAIN"),
			utils.GetEnv("SECURITY_PROTOCOL", "SASL_SSL"),
			utils.GetEnv("SASL_USERNAME", ""),
			utils.GetEnv("SASL_PASSWORD", ""),
		)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		m.Broadcast(msg)
	})

	httpPort := utils.GetEnv("HTTP_PORT", "8080")
	log.Debugf("Listening on %v...", httpPort)

	return r, httpPort
}
