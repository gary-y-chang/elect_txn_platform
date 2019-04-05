package rest

import (
	//"github.com/gorilla/sessions"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var Router *echo.Echo
var RedisConn redis.Conn

//var store *sessions.CookieStore

func init() {
	Router = echo.New()

	var err error

	RedisConn, err = redis.Dial("tcp", "redis:6379")
	//RedisConn, err = redis.Dial("tcp", "192.168.1.4:6379")
	if err != nil {
		panic(err)
	}

	// store = sessions.NewCookieStore([]byte("session-secret"))
	// store.Options = &sessions.Options{
	// 	HttpOnly: true,
	// 	Secure:   false,
	// }

	// Middleware
	Router.Use(middleware.Logger())
	Router.Use(middleware.Recover())

	Router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "Cache-Control"},
	}))
	//Router.Static("/static", "rest")
	Router.Static("/static", "/goapp")

	Router.GET("/", home) // for test

	Router.POST("/login", login)
	Router.POST("/powerrec/add", addPowerRecord)

	adm := Router.Group("/admin")
	adm.Use(middleware.JWT([]byte("secret")))
	adm.GET("/users/all", getAllUsers)
	adm.POST("/users/add", createUser)
	adm.POST("/meters/add", createMeterDeposit)

	g := Router.Group("/platform")
	g.Use(middleware.JWT([]byte("secret")))

	g.GET("/user/precords/:meter_id", getPowerRecordsByMID)

	g.GET("/user/dashboard/:uid", getDashboardInfo)
	g.GET("/user/meters/:uid", getUserMeters)
	g.GET("/user/balance/:deposit_no", getDepositBalance)
	g.PATCH("/user/balance/increase", addValueToBalance) // application/json  {DepositNo: xxxxxx, Balance: 100.0}
	//g.PUT("/user/meter/switch", switchMeterInUse)      // application/json  models.MeterDeposit
	g.PATCH("/user/meter/switch", switchMeterInUse) // application/json {MeterID: xxxxxx}

	g.GET("/order/:type/:meter_id", getUndealtOrdersByMID)

	g.POST("/order/add", createOrder)
}
