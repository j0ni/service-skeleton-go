package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	network "github.com/glibs/gin-webserver"
	olioMiddleware "github.com/olioapps/service-skeleton-go/olio/service/middleware"
	olioResources "github.com/olioapps/service-skeleton-go/olio/service/resources"
	"github.com/olioapps/service-skeleton-go/olio/util"
)

type OlioDaemon interface {
	Start()
	Stop()
}

type OlioResourceHandler interface {
	Init(*gin.Engine, *olioMiddleware.WhiteList)
}

type OlioBaseService struct {
	GinEngine       *gin.Engine
	server          *network.WebServer
	daemons         []OlioDaemon
	versionResource *olioResources.VersionResource
	healthResource  *olioResources.HealthResource
}

func New() *OlioBaseService {
	service := OlioBaseService{}
	service.GinEngine = gin.Default()

	return &service
}

func (obs *OlioBaseService) Init(whitelist *olioMiddleware.WhiteList, middlewares []gin.HandlerFunc, resources []OlioResourceHandler) {
	log.Info("Initializing RESTful service.")

	log.Debug("Setting up middleware.")

	obs.GinEngine.Use(whitelist.Handler)
	obs.GinEngine.Use(olioMiddleware.RequestId)

	for _, middleware := range middlewares {
		obs.GinEngine.Use(middleware)
	}

	healthResource := olioResources.NewHealthResource()
	obs.healthResource = healthResource
	healthResource.Init(obs.GinEngine)

	versionResource := olioResources.NewVersionResource()
	obs.versionResource = versionResource
	versionResource.Init(obs.GinEngine)

	pingResource := olioResources.NewPingResource()
	pingResource.Init(obs.GinEngine, whitelist)
	for _, resource := range resources {
		resource.Init(obs.GinEngine, whitelist)
	}

	log.Debug("Setting up routes.")
}

func (obs *OlioBaseService) AddDaemon(daemon OlioDaemon) {
	obs.daemons = append(obs.daemons, daemon)
}

func (obs *OlioBaseService) AddVersionExtractor(versionExtractor olioResources.VersionExtractor) {
	obs.versionResource.AddVersionExtractor(versionExtractor)
}

func (obs *OlioBaseService) AddUptimeExtractor(uptimeExtractor olioResources.UptimeExtractor) {
	obs.healthResource.AddUptimeExtractor(uptimeExtractor)
}

func (obs *OlioBaseService) Start() {
	for _, daemon := range obs.daemons {
		daemon.Start()
	}

	servicePort := util.GetEnv("PORT", "9090")
	host := ":" + servicePort
	log.Info("Starting webserver at PORT ", servicePort)
	obs.server = network.InitializeWebServer(obs.GinEngine, host)
	obs.server.Start()
}

func (obs *OlioBaseService) Stop() {
	for _, daemon := range obs.daemons {
		daemon.Stop()
	}
	obs.server.Stop()
}
