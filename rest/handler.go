package rest

import (
	"github.com/labstack/echo"
	"gitlab.com/wondervoyage/platform/models"
	"net/http"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
	"strconv"
	"log"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/satori/go.uuid"
)

func home(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome Home !")
}

func fabricSdk(c echo.Context) error {
	configProvider := config.FromFile("./configs/fabric/config.yaml")
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		log.Fatalf("create sdk fail: %s\n", err.Error())
	}
	defer  sdk.Close()

	//读取配置文件(config.yaml)中的组织(member1.example.com)的用户(Admin)
	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg("Org1"))
	if err != nil {
		log.Fatalf("create msp client fail: %s\n", err.Error())
	}

	adminIdentity, err := mspClient.GetSigningIdentity("Admin")
	if err != nil {
		log.Fatalf("get admin identify fail: %s\n", err.Error())
	} else {
		fmt.Println("Admin Identify is found:")
		fmt.Println(adminIdentity)
	}

	//调用合约
	channelProvider := sdk.ChannelContext("mychannel",
		fabsdk.WithUser("User1"),
		fabsdk.WithOrg("Org1"))

	channelClient, err := channel.New(channelProvider)
	//_, err := channel.New(channelProvider)
	if err != nil {
		log.Fatalf("create channel client fail: %s\n", err.Error())
	} else {
		fmt.Println("channelClient create successful !!!")
	}

	/****** query operation  ------------------*/
	var args [][]byte
	args = append(args, []byte("b"), []byte("a"), []byte("40"))

	request := channel.Request{
		ChaincodeID: "example02",
		Fcn:         "invoke",
		Args:        args,
	}

	response, err := channelClient.Execute(request)
	if err != nil {
		log.Fatal("query fail: ", err.Error())
	} else {
		fmt.Printf("response is %s\n", response.Payload)
	}

	return c.String(http.StatusOK, "Welcome to Hyperledger !")
}

func createUser(c echo.Context) error {
	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	if err :=models.AddUser(*u); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}

func createOrder(c echo.Context) error {
	ord := new(models.Order)
	if err := c.Bind(ord); err != nil {
		return err
	}
	//tp := c.Param("type")  //1:buy, 2:sell
	switch tp := c.Param("type") ; tp {
		case "buy":
			ord.Type = 1
		case "sell":
			ord.Type = 2
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "URL parameter error, order/[buy/sell]")
			//return errors.New("request parameter error, correct:[buy/sell]")
	}
	ord.ID = uuid.Must(uuid.NewV4()).String()
	ord.Status = 1
	
	if err :=models.AddOrder(*ord); err != nil {
		return err
	}

	//TODO  next, invoke the chaincode to place the order


	return c.JSON(http.StatusOK, ord)
}

func getAllUsers(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	account := claims["account"].(string)
	isAdmin := claims["admin"].(bool)

	fmt.Printf("Request from user: %s [admin=%t]", account, isAdmin)

    users := models.AllUsers()
	return c.JSON(http.StatusOK, users)
}

func getPowerRecordsByUID(c echo.Context) error {
	userId := c.Param("uid")
	uid, err := strconv.Atoi(userId); if err != nil {
		return err
	}
	records := models.GetUserPowerRecords(uint(uid))
	return c.JSON(http.StatusOK, records)
}

func getUndealtOrdersByUID(c echo.Context) error {
	//userId := c.QueryParam("uid")
	//status := c.QueryParam("state")
	var tpe  uint8
	uid, err := strconv.Atoi(c.Param("uid")); if err != nil {
		return err
	}

	switch tp := c.Param("type") ; tp {
	case "buy":
		tpe = 1
	case "sell":
		tpe = 2
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "URL parameter error, order/[buy/sell]/[uid]")
		//return errors.New("request parameter error, correct:[buy/sell]")
	}

	orders := models.GetUserUndealtOrders(uint(uid), tpe)
	return c.JSON(http.StatusOK, orders)
}

func login(c echo.Context) error {
	account := c.FormValue("account")
	pass := c.FormValue("password")
	fmt.Printf("Login Info: %s | %s\n", account, pass)

	u, logged := models.LoginCheck(account, pass)
	fmt.Printf("Logged User : %s\n", u.Account)

	if logged {
		// Create token
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["account"] = u.Account
		claims["uid"] = u.ID
		claims["admin"] = false
		claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte("secret"))
		if err != nil {
			return err
		}

		//userByte, err := json.Marshal(u); if err != nil {
		//	return err
		//}

		return c.JSON(http.StatusOK, map[string]string{
			"user_id": strconv.Itoa(int(u.Model.ID)),
			"user_account" : u.Account,
			"token": t,
		})
	}

	return echo.ErrUnauthorized
}
