package middleware

import (
	"net/http"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/j0ni/service-skeleton-go/olio/api"
	"github.com/j0ni/service-skeleton-go/olio/util"
)

type OlioJWTAuthMiddleware struct {
	userExtractor  UserExtractor
	tokenValidator api.TokenValidator
}

func NewOlioJWTAuthMiddleware(userExtractor UserExtractor, tokenValidator api.TokenValidator) OlioJWTAuthMiddleware {
	middleware := OlioJWTAuthMiddleware{userExtractor, tokenValidator}

	return middleware
}

func (m OlioJWTAuthMiddleware) Create() gin.HandlerFunc {

	return func(c *gin.Context) {
		// if the user exists, continue
		_, exists := c.Get("whitelisted")
		if exists {
			c.Next()
			return
		}

		user, exists := c.Get("currentUser")
		if exists && user != nil {
			c.Next()
			return
		}

		authHeader := c.Request.Header.Get("Authorization")

		// skip this handler if no authorization header is found
		if authHeader == "" {
			c.Next()
			return
		}

		token := util.RequestToToken(c)
		if token == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		blacklisted, exception := m.tokenValidator.IsTokenBlacklisted(token.Raw)
		if exception != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if blacklisted {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("JWT_TOKEN", token.Raw)
		var f map[string]interface{} = token.Claims
		username, ok := f["subject"].(string)
		if ok {
			requestID, _ := c.Get("Request-Id")
			user, err2 := m.userExtractor.ExtractUserByUsername(username, requestID.(string))
			if err2 != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			authorizedResources, ok := f["authorizedResources"]
			if !ok && authorizedResources == nil {
				c.Set("currentUser", user)
			} else {
				// analyze authorized resources
				authorizedResourcesParts := strings.Split(authorizedResources.(string), ":")
				method := authorizedResourcesParts[0]
				path := authorizedResourcesParts[1]

				if path != c.Request.URL.Path || (method != "*" && method != c.Request.Method) {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				} else {
					c.Set("currentUser", user)
				}

			}
		}

		c.Next()
	}

}
