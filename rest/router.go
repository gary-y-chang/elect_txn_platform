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

	Router.GET("/", home)
	Router.POST("/login", login)
	Router.POST("user/add", createUser)

	g := Router.Group("/platform")
	g.Use(middleware.JWT([]byte("secret")))
	g.GET("/user/all", getAllUsers)
    g.GET("/user/precords/:uid", getPowerRecordsByUID)
}






