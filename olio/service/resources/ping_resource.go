package resources

import (
	"github.com/gin-gonic/gin"
	"github.com/j0ni/service-skeleton-go/olio/service/middleware"
	log "github.com/sirupsen/logrus"
)

type PingResource struct {
}

func NewPingResource() PingResource {
	obj := PingResource{}

	return obj
}

func (resource PingResource) Init(r *gin.Engine, whiteList *middleware.WhiteList) {
	log.Debug("Setting up ping resource.")

	r.GET("/api/ping", resource.ping)
}

func (resource PingResource) ping(c *gin.Context) {
	c.Writer.WriteString("pong")
}
