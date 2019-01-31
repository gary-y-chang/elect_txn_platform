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

	Router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, " api_key"},
	}))
	Router.Static("/static", "rest")
	//Router.Static("/static", "/goapp")

	Router.GET("/", home)  // for test
	Router.GET("/fabric", fabricSdk) // for test
	Router.POST("/login", login)
	Router.POST("/user/add", createUser)


	g := Router.Group("/platform")
	g.Use(middleware.JWT([]byte("secret")))
	g.GET("/user/all", getAllUsers)
    g.GET("/user/precords/:uid", getPowerRecordsByUID)
	g.GET("/user/dashboard/:uid", getDashboardInfo)
	g.GET("/order/:type/:uid", getUndealtOrdersByUID)
	g.POST("/order/add", createOrder)
}






