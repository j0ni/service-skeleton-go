package resources

import (
	"github.com/gin-gonic/gin"
	"github.com/j0ni/service-skeleton-go/olio/extractors"
	"github.com/j0ni/service-skeleton-go/olio/models"
	log "github.com/sirupsen/logrus"
)

type VersionResource struct {
	BaseResource
	versionExtractor extractors.VersionExtractor
}

const VERSION = "1.0.4"

func NewVersionResource() *VersionResource {
	obj := VersionResource{}

	return &obj
}

func (vr *VersionResource) AddVersionExtractor(versionExtractor extractors.VersionExtractor) {
	vr.versionExtractor = versionExtractor
}

func (vr *VersionResource) Init(r *gin.Engine) {
	log.Debug("Setting up version resource.")

	r.GET("/api/version", vr.getVersion)
}

func (vr *VersionResource) getVersion(c *gin.Context) {
	skeletonVersion := VERSION
	var appVersion string
	if vr.versionExtractor != nil {
		appVersion = vr.versionExtractor.GetAppName() + "-" + vr.versionExtractor.GetVersion()
	} else {
		appVersion = "no version given"
	}

	version := models.Version{
		AppVersion:      appVersion,
		SkeletonVersion: skeletonVersion,
	}

	vr.ReturnJSON(c, 200, version)
}
