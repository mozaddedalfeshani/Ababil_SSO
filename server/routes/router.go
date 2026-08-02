package routes

import (
	"ababilx-sso/middleware"

	"github.com/gin-gonic/gin"
)

// NewEngine builds the Gin engine with trusted-proxy configuration
// locked to the given CIDRs. This is the single place that decision is
// made: if it were left to Gin's default (trust everyone), any client
// could set X-Forwarded-For and forge its way past every per-IP rate
// limit and lockout.
func NewEngine(trustedProxyCIDRs []string, isProduction bool) (*gin.Engine, error) {
	if isProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	if err := r.SetTrustedProxies(trustedProxyCIDRs); err != nil {
		return nil, err
	}

	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogging())
	r.Use(middleware.SecurityHeaders())

	return r, nil
}
