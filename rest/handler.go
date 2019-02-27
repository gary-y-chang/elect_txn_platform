package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/wondervoyage/platform/models"
	"gitlab.com/wondervoyage/platform/simulation"
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
	defer sdk.Close()

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
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	account := claims["account"].(string)
	isAdmin := claims["admin"].(bool)
	fmt.Printf("Request from user: %s [admin=%t]\n", account, isAdmin)
	// check if isAdmin not true, return "Authorized error"

	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	if err := models.AddUser(*u); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}

func createMeterDeposit(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	account := claims["account"].(string)
	isAdmin := claims["admin"].(bool)
	fmt.Printf("Request from user: %s [admin=%t]\n", account, isAdmin)
	// check if isAdmin not true, return "Authorized error"

	meter := new(models.MeterDeposit)
	if err := c.Bind(meter); err != nil {
		return err
	}

	// meter.DepositNo rule  xxxx-xx-xxxxxxx
	// section 1: 國碼+OrgID, e.g 8863(org_id=3) 88615(org_id=15)
	// section 2: UserID is string
	// section 3: 固定7碼, starting from 1, 流水號累加，前面補0, e.g. 0000018 or 0000123
	//s := strings.Split("127.0.0.1:5432", ":")
	//ip, port := s[0], s[1]
	var sec3 string
	prevMeter := models.GetLatestMeterDeposit(meter.OrgID)
	if prevMeter == nil {
		sec3 = "0000001"
	} else {
		prevNo := prevMeter.(models.MeterDeposit).DepositNo
		s := strings.Split(prevNo, "-") //s[0] s[1] s[2]
		i, _ := strconv.Atoi(strings.Trim(s[2], "0"))
		sec3 = fmt.Sprintf("%07d", i+1)
	}

	sec1 := "886" + strconv.Itoa(int(meter.OrgID))
	sec2 := strconv.Itoa(int(meter.UserID))
	sec := []string{sec1, sec2, sec3}
	meter.DepositNo = strings.Join(sec, "-")

	local, _ := time.LoadLocation("Local")
	now := time.Now().In(local)
	meter.CreatedAt = now
	meter.UpdatedAt = now
	//if err := models.AddMeterDeposit(*meter); err != nil {
	//	return err
	//}
	// the invoke chaincode to create the Deposit

	//prepare  Deposit
	deposit := models.Deposit{meter.DepositNo,
		0, 0, now, now, meter.UserID}

	//prepare  DepositRecord
	rec := models.DepositRecord{deposit.DepositNo, deposit.UserID,
		"create new deposit", 0, 0,
		0, "Meter Id: " + meter.MeterID, now}

	jsonDepo, _ := json.Marshal(deposit)
	//TODO  invoke chaincode to create the user's Deposit of this meter
	_, err := simulation.Invoke("create", []string{string(jsonDepo)})
	if err == nil {
		if dberr := models.AddMeterDeposit(*meter); dberr != nil {
			return dberr
		}
		if dberr := models.AddDepositRecord(rec); dberr != nil {
			return dberr
		}
	} else {
		return c.JSON(http.StatusExpectationFailed, err)
	}

	return c.JSON(http.StatusOK, meter)
}

func createOrder(c echo.Context) error {
	ord := new(models.Order)
	if err := c.Bind(ord); err != nil {
		return err
	}

	//get the user's selected deposit_no, then assign it to  ord.DepositNo
	ord.DepositNo = models.GetSelectedMeterDeposit(ord.UserID).DepositNo
	//ord.ID = uuid.Must(uuid.NewV4()).String()
	ord.ID = uuid.NewV4().String()

	//TODO:  next, invoke the chaincode to place the order,  a DealTxn should be returned
	// here is a simulation
	var txnJson []byte
	var err error
	if ord.Type == 1 { //buy
		//TODO:  invoke chaincode to check the user's balance see if enough balance to pay this order
		args := []string{ord.DepositNo}
		byteDeposit, _ := simulation.Invoke("query", args)

		var depo models.Deposit
		json.Unmarshal(byteDeposit, &depo)
		if depo.Balance < ord.Price*ord.Kwh {
			return errors.New("no enough balance to buy")
		} // if no, return error with "Balance not enough" message
		//TODO:  invoke chaincode
		txnJson, err = simulation.BuyHandler(*ord)
		if err != nil {
			return err
		} else {
			// if txn succeeds, invoke chaincode balance.go with function detain to detain payable
			//bt := "{\"Target\": \"" + ord.DepositNo + "\", \"Source\": \"\", \"Amount\": " + fmt.Sprintf("%.2f",ord.Price*ord.Kwh) + "}"
			//args := []string{bt}
			////TODO  invoke chaincode
			//simulation.Invoke("detain", args)
			fmt.Printf("======> before persist buy: %+v\n", ord)
			err = models.AddOrder(*ord)
			if err != nil {
				return err
			}
		}
	} else if ord.Type == 2 { //sell
		//  check the user's saleable power see if enough saleable power to sell
		//if no, return error with "Stocked Power not enough." message
		power := models.GetLatestPowerRecord(ord.UserID)
		if power.KwhSaleable < ord.Kwh {
			return errors.New("no enough power for sale")
		}
		//TODO  invoke chaincode
		txnJson, err = simulation.SellHandler(*ord)
		if err != nil {
			return err
		} else {
			fmt.Printf("======> before persist sell: %+v\n", ord)
			if err := models.AddOrder(*ord); err != nil {
				return err
			}
		}
	}

	var txns []models.DealTxn
	json.Unmarshal(txnJson, &txns)
	for i, t := range txns {
		if t.Kwh > 0 { // if a deal complete, insert this TXN into DB, then update the buy and sell orders
			fmt.Println("index: " + strconv.Itoa(i))
			t.Part = uint8(i + 1)
			txns[i] = t
			models.AddDealTxn(t)
			models.UpdateOrderByTxn(t)
			models.UpdatePowerRecordByTxn(t)
			bt := "{\"Target\": \"" + t.SellDepositNo + "\", \"Source\": \"" + t.BuyDepositNo +
				"\", \"Amount\": " + fmt.Sprintf("%.2f", t.Kwh*t.Price) + "}"
			args := []string{bt}

			//TODO  invoke chaincode
			byteDepos, err := simulation.Invoke("transfer", args)
			if err != nil {
				return err
			} else { // writing the DepositRecord
				var depos []models.Deposit //depos[0] ->from   depos[1] ->to
				json.Unmarshal(byteDepos, &depos)
				fromRecord := models.DepositRecord{depos[0].DepositNo, depos[0].UserID,
					"buy order expense", -(t.Kwh * t.Price), depos[0].Balance,
					0, "Order Id: " + t.BuyOrderID, t.TxnDate}
				models.AddDepositRecord(fromRecord)

				toRecord := models.DepositRecord{depos[1].DepositNo, depos[1].UserID,
					"sell order income", t.Kwh * t.Price, depos[1].Balance,
					0, "Order Id: " + t.SellOrderID, t.TxnDate}
				models.AddDepositRecord(toRecord)

			}
		}
	}

	return c.JSON(http.StatusOK, txns)
}

func getAllUsers(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	account := claims["account"].(string)
	isAdmin := claims["admin"].(bool)
	fmt.Printf("Request from user: %s [admin=%t]\n", account, isAdmin)
	// check if isAdmin not true, return "Authorized error"

	pp, _ := strconv.Atoi(c.QueryParam("page"))
	users, count := models.AllUsers(pp, 10)
	// fmt.Printf("%+v\n", users)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total": count,
		"users": users,
	})
}

func getPowerRecordsByUID(c echo.Context) error {
	userId := c.Param("uid")
	uid, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	pp, _ := strconv.Atoi(c.QueryParam("page"))
	records, count := models.GetUserPowerRecords(uint(uid), pp, 10)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"total":   count,
		"records": records,
	})
}

func getUndealtOrdersByUID(c echo.Context) error {
	//userId := c.QueryParam("uid")
	//status := c.QueryParam("state")
	var tpe uint8
	uid, err := strconv.Atoi(c.Param("uid"))
	if err != nil {
		return err
	}

	switch tp := c.Param("type"); tp {
	case "buy":
		tpe = 1
	case "sell":
		tpe = 2
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "URL parameter error, order/[buy/sell]/[uid]")
		//return errors.New("request parameter error, correct:[buy/sell]")
	}

	pp, _ := strconv.Atoi(c.QueryParam("page"))
	orders, count := models.GetUserUndealtOrders(uint(uid), tpe, pp, 10)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"total":  count,
		"orders": orders,
	})
}

func getDashboardInfo(c echo.Context) error {
	uid, err := strconv.Atoi(c.Param("uid"))
	if err != nil {
		return err
	}
	var buyKwh float64
	buyOrders := models.GetUserAllUndealtOrders(uint(uid), 1)
	for _, bo := range buyOrders {
		buyKwh += bo.Kwh
	}
	var sellKwh float64
	sellOrders := models.GetUserAllUndealtOrders(uint(uid), 2)
	for _, so := range sellOrders {
		sellKwh += so.Kwh
	}

	prd := models.GetLatestPowerRecord(uint(uid))

	return c.JSON(http.StatusOK, map[string]string{
		"stocked":  fmt.Sprintf("%.2f", prd.KwhStocked),
		"saleable": fmt.Sprintf("%.2f", prd.KwhSaleable),
		"on_sell":  fmt.Sprintf("%.2f", sellKwh),
		"on_buy":   fmt.Sprintf("%.2f", buyKwh),
	})
}

//TODO:
func getUserMeters(c echo.Context) error {
	userId := c.Param("uid")
	uid, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	meters := models.GetUserMeters(uint(uid))

	return c.JSON(http.StatusOK, meters)
}

func getDepositBalance(c echo.Context) error {
	depositNo := c.Param("deposit_no")
	args := []string{depositNo}

	//TODO: invoke chaincode
	byteDeposit, err := simulation.Invoke("query", args)
	if err != nil {
		return err
	}
	var depo models.Deposit
	json.Unmarshal(byteDeposit, &depo)

	return c.JSON(http.StatusOK, depo)
}

//TODO:
func addValueToBalance(c echo.Context) error {
	depo := new(models.Deposit)
	if err := c.Bind(depo); err != nil {
		return err
	}
	//bt := "{\"Target\": \""+ 8867-3-0000002+"\", \"Source\": \"\", \"Amount\": 200}"
	bTxn := models.BalanceTxn{depo.DepositNo, "", depo.Balance}
	byteBT, _ := json.Marshal(bTxn)
	args := []string{string(byteBT)}
	byteDeposit, err := simulation.Invoke("deposit", args)
	if err != nil {
		return err
	}

	json.Unmarshal(byteDeposit, depo)

	return c.JSON(http.StatusOK, depo)
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
		claims["exp"] = time.Now().Add(time.Minute * 120).Unix()

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte("secret"))
		if err != nil {
			return err
		}

		c.Response().Header().Set("Cache-Control", "no-cache")
		return c.JSON(http.StatusOK, map[string]string{
			"user_id":      strconv.Itoa(int(u.Model.ID)),
			"user_account": u.Account,
			"token":        t,
		})
	}

	return echo.ErrUnauthorized
}
