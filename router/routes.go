package router

import "github.com/labstack/echo/v4"

func Routes(e *echo.Echo) {

	v1Routes := e.Group("/api/v1")
	v1Apis(v1Routes)
}
