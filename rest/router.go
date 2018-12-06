package rest

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var Router *echo.Echo
func init() {
	Router = echo.New()

	// Middleware
	Router.Use(middleware.Logger())
	Router.Use(middleware.Recover())
}


