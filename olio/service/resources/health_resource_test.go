package resources

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/j0ni/service-skeleton-go/olio/extractors"
	"github.com/j0ni/service-skeleton-go/olio/util"
)

type UptimeExtractor struct {
}

func (ue *UptimeExtractor) GetUptime() time.Duration {
	return 5000
}

type DbExtractor struct {
}

func (de DbExtractor) ExtractDialect() extractors.DbDialect {
	return extractors.Postgres
}

func (de DbExtractor) ExtractConnectionString() string {
	return util.GetEnv("DB_CONNECTION_STRING", "")
}

type DbHealthExtractor struct {
}

func (de *DbHealthExtractor) GetDbExtractor() extractors.DbExtractor {
	return DbExtractor{}
}

func TestHealth(t *testing.T) {
	uptimeExtractor := UptimeExtractor{}
	dbHealthExtractor := DbHealthExtractor{}

	tt := []struct {
		name     string
		uptime   *UptimeExtractor
		dbHealth *DbHealthExtractor
	}{
		{name: "just uptime", uptime: &uptimeExtractor, dbHealth: nil},
		{name: "uptime and dbHealth", uptime: &uptimeExtractor, dbHealth: &dbHealthExtractor},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			healthResource := NewHealthResource()
			healthResource.AddUptimeExtractor(tc.uptime)
			if tc.name == "uptime and dbHealth" {
				healthResource.AddDbHealthExtractor(tc.dbHealth)
			}

			router.GET("/api/health", healthResource.getHealth)
			req, _ := http.NewRequest("GET", "/api/health", nil)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			if tc.name == "uptime and dbHealth" {
				if util.GetEnv("DB_CONNECTION_STRING", "") == "" {
					assert.Equal(t, "{\n    \"uptime\": \"0.001 hours\",\n    \"dbOk\": \"false\"\n}", res.Body.String())
					return
				}
			}

			assert.Equal(t, "{\n    \"uptime\": \"0.001 hours\"\n}", res.Body.String())
		})
	}
}
