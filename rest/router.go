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
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "Cache-Control"},
	}))
	Router.Static("/static", "rest")
	//Router.Static("/static", "/goapp")

	Router.GET("/", home)            // for test
	Router.GET("/fabric", fabricSdk) // for test
	Router.POST("/login", login)
	//Router.POST("/user/add", createUser)

	adm := Router.Group("/admin")
	adm.Use(middleware.JWT([]byte("secret")))
	adm.GET("/users/all", getAllUsers)
	adm.POST("/users/add", createUser)
	adm.POST("/meters/add", createMeterDeposit)

	g := Router.Group("/platform")
	g.Use(middleware.JWT([]byte("secret")))
	//g.GET("/user/all", getAllUsers)
	g.GET("/user/precords/:uid", getPowerRecordsByUID)
	g.GET("/user/dashboard/:uid", getDashboardInfo)
	g.GET("/user/meters/:uid", getUserMeters)
	g.GET("/user/balance/:deposit_no", getDepositBalance)
	g.PUT("/user/balance/increase", addValueToBalance) // application/json  models.Deposit

	g.GET("/order/:type/:uid", getUndealtOrdersByUID)
	g.POST("/order/add", createOrder)
}
