package server

import (
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"strings"
)

func New() *echo.Echo {
	echoInstance := echo.New()
	p := prometheus.NewPrometheus("echo", urlSkipper)
	p.Use(echoInstance)

	echoInstance.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		// Skipping logging for health checking api
		Skipper: func(c echo.Context) bool {
			if c.Request().RequestURI == "/health" || c.Request().RequestURI == "/metrics" {
				return true
			}
			return false
		},
		Format: "[${time_rfc3339}] method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
	echoInstance.Use(middleware.Recover())
	return echoInstance
}

func urlSkipper(c echo.Context) bool {
	if strings.HasPrefix(c.Path(), "/index") {
		return true
	} else if strings.HasPrefix(c.Path(), "/health") {
		return true
	}
	return false
}
