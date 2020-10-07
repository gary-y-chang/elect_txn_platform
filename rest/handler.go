package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"gitlab.com/wondervoyage/platform/models"
	"gitlab.com/wondervoyage/platform/simulation"
)

type orderQuery struct {
	MeterID string `json:"meter_id"`
	Tpe     uint8  `json:"type"`
	Status  uint8  `json:"status"`
	Begin   string `json:"begin"`
	End     string `json:"end"`
}

type powerQuery struct {
	MeterID string `json:"meter_id"`
	Begin   string `json:"begin"`
	End     string `json:"end"`
}

func home(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome Home !")
}

func createUser(c echo.Context) error {
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	account := claims["account"].(string)
	isAdmin := claims["admin"].(bool)
	fmt.Printf("Request from user: %s [admin=%t]\n", account, isAdmin)
	// check if isAdmin not true, return "Authorized error"

	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	if err := u.AddUser(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, u)
}

func createMeterDeposit(c echo.Context) error {
	//TODO: logic for doing authorization on 'isAdmin'
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
		i, _ := strconv.Atoi(strings.TrimLeft(s[2], "0"))
		sec3 = fmt.Sprintf("%07d", i+1)
	}

	sec1 := "886" + strconv.Itoa(int(meter.OrgID))
	sec2 := strconv.Itoa(int(meter.UserID))
	sec := []string{sec1, sec2, sec3}
	meter.DepositNo = strings.Join(sec, "-")

	now := currentLocalTime()
	meter.CreatedAt = now
	meter.UpdatedAt = now

	//prepare  Deposit
	deposit := models.Deposit{meter.DepositNo, 0, now, now, meter.UserID}

	//prepare  DepositRecord
	rec := models.DepositRecord{deposit.DepositNo, deposit.UserID,
		"create new deposit", 0, 0,
		0, "Meter Id: " + meter.MeterID, now}

	jsonDepo, _ := json.Marshal(deposit)
	//TODO:  invoke chaincode to create the user's Deposit of this meter
	_, err := simulation.Invoke("create", []string{string(jsonDepo)})
	//_, err := chaincaller.Balance("create", string(jsonDepo))
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
	account := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["account"].(string)
	//user := c.Get("user").(*jwt.Token)
	//claims := user.Claims.(jwt.MapClaims)
	//account := claims["account"].(string)

	ord.DepositNo = meterInUse[account].DepositNo
	ord.MeterID = meterInUse[account].MeterID
	ord.Status = 1
	ord.ID = uuid.NewV4().String()

	//TODO:  next, invoke the chaincode to place the order,  a DealTxn should be returned
	var txnJson []byte
	var err error
	if ord.Type == 1 { //buy

		args := []string{ord.DepositNo}
		byteDeposit, _ := simulation.Invoke("query", args)
		//TODO:  invoke chaincode to check the user's balance see if enough balance to pay this order
		//byteDeposit, _ := chaincaller.Balance("query", ord.DepositNo)

		var depo models.Deposit
		json.Unmarshal(byteDeposit, &depo)
		payable := models.GetPayableByMeter(ord.MeterID)
		if (depo.Balance - payable) < ord.Price*ord.Kwh {
			return c.String(http.StatusInternalServerError, "no enough balance to buy")
		}
		//TODO:  invoke chaincode
		txnJson, err = simulation.BuyHandler(*ord)
		//txnJson, err = chaincaller.PlaceOrder("buy", *ord)
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
		reserved := models.GetReservedKwhByMeter(ord.MeterID)
		power := models.GetLatestPowerSaleable(ord.MeterID)
		if (power.KwhSaleable - reserved) < ord.Kwh {
			return c.String(http.StatusInternalServerError, "no enough power for sale")
		}
		//TODO: invoke chaincode
		txnJson, err = simulation.SellHandler(*ord)
		//txnJson, err = chaincaller.PlaceOrder("sell", *ord)
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
			byteDepos, err := simulation.Invoke("transfer", args)
			//TODO:  invoke chaincode
			//byteDepos, err := chaincaller.Balance("transfer", bt)

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

func getPowerAnalysis(c echo.Context) error {
	// meterId := c.Param("meter_id")
	// uid, err := strconv.Atoi(userId)
	// if err != nil {
	// 	return err
	// }
	q := new(powerQuery)
	if err := c.Bind(q); err != nil {
		return err
	}
	format := "2006-01-02"
	begin, _ := time.Parse(format, q.Begin)
	end, _ := time.Parse(format, q.End)
	pp, _ := strconv.Atoi(c.QueryParam("page"))

	analysis, count := models.GetPowerAnalysis(q.MeterID, begin, end, pp, 10)
	fmt.Printf("total %d -------------> %+v \n", count, analysis)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"total":   count,
		"summary": analysis,
	})
	// records, count := models.GetPowerRecordsByMeter(meterId, pp, 10)
	// return c.JSON(http.StatusOK, map[string]interface{}{
	// 	"total":   count,
	// 	"records": records,
	// })
}

func getOrdersByCondition(c echo.Context) error {
	// tpe, _ := strconv.Atoi(c.FormValue("type"))
	// status, _ := strconv.Atoi(c.FormValue("status"))

	// format := "2006-01-02 15:04:05"
	// begin, _ := time.Parse(format, c.FormValue("begin"))
	// end, _ := time.Parse(format, c.FormValue("end"))
	//account := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["account"].(string)
	q := new(orderQuery)
	if err := c.Bind(q); err != nil {
		return err
	}
	format := "2006-01-02"
	begin, _ := time.Parse(format, q.Begin)
	end, _ := time.Parse(format, q.End)
	pp, _ := strconv.Atoi(c.QueryParam("page"))

	//orders, count := models.QueryOrders(uint8(tpe), uint8(status), begin, end, pp, 10)
	orders, count := models.QueryOrders(q.Tpe, q.Status, q.MeterID, begin, end, pp, 10)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total":  count,
		"orders": orders,
	})
}

func getUndealtOrdersByMID(c echo.Context) error {
	meterId := c.Param("meter_id")

	var tpe uint8
	switch tp := c.Param("type"); tp {
	case "buy":
		tpe = 1
	case "sell":
		tpe = 2
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "URL parameter error, order/[buy/sell]/[meter_id]")
		//return errors.New("request parameter error, correct:[buy/sell]")
	}

	pp, _ := strconv.Atoi(c.QueryParam("page"))
	orders, count := models.GetUndealtOrdersByMeter(meterId, tpe, pp, 10)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"total":  count,
		"orders": orders,
	})
}

func getDashboardInfo(c echo.Context) error {
	account := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["account"].(string)
	userid := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["uid"].(float64)
	//user := c.Get("user").(*jwt.Token)
	//claims := user.Claims.(jwt.MapClaims)
	//account := claims["account"].(string)
	if c.Param("uid") != fmt.Sprintf("%g", userid) {
		return c.String(http.StatusBadRequest, "rest path parameter uid not matching token user id")
	}

	meter := meterInUse[account]
	args := []string{meter.DepositNo}
	byteDeposit, err := simulation.Invoke("query", args)
	//TODO: invoke chaincode
	//byteDeposit, err := chaincaller.Balance("query", meter.DepositNo)
	if err != nil {
		return err
	}
	var depo models.Deposit
	json.Unmarshal(byteDeposit, &depo)

	// uid, err := strconv.Atoi(c.Param("uid"))
	// if err != nil {
	// 	return err
	// }

	buyKwh := models.GetUndealtKwhByMeter(meter.MeterID, 1)
	sellKwh := models.GetUndealtKwhByMeter(meter.MeterID, 2)
	saleable := models.GetLatestPowerSaleable(meterInUse[account].MeterID)
	stock := models.GetLatestPowerStocked(meterInUse[account].MeterID)

	return c.JSON(http.StatusOK, map[string]string{
		"saleable":   fmt.Sprintf("%.2f", saleable.KwhSaleable),
		"stocked":    fmt.Sprintf("%.2f", stock.KwhStocked),
		"on_sell":    fmt.Sprintf("%.2f", sellKwh),
		"on_buy":     fmt.Sprintf("%.2f", buyKwh),
		"meter_id":   meter.MeterID,
		"meter_name": meter.MeterName,
		"deposit_no": meter.DepositNo,
		"balance":    fmt.Sprintf("%.1f", depo.Balance)})
}

func getUserMeters(c echo.Context) error {
	usrid := c.Param("uid")
	uid, err := strconv.Atoi(usrid)
	if err != nil {
		return err
	}
	meters := models.GetUserMeters(uint(uid))

	return c.JSON(http.StatusOK, meters)
}

func getDepositBalance(c echo.Context) error {
	depositNo := c.Param("deposit_no")
	args := []string{depositNo}
	byteDeposit, err := simulation.Invoke("query", args)
	//TODO: invoke chaincode
	//byteDeposit, err := chaincaller.Balance("query", depositNo)
	if err != nil {
		return err
	}
	var depo models.Deposit
	json.Unmarshal(byteDeposit, &depo)

	return c.JSON(http.StatusOK, depo)
}

func addValueToBalance(c echo.Context) error {
	depo := new(models.Deposit)
	if err := c.Bind(depo); err != nil {
		return err
	}
	//bt := "{\"Target\": \""+ 8867-3-0000002+"\", \"Source\": \"\", \"Amount\": 200}"
	plus := depo.Balance
	bTxn := models.BalanceTxn{depo.DepositNo, "", plus}
	byteBT, _ := json.Marshal(bTxn)
	args := []string{string(byteBT)}
	byteDeposit, err := simulation.Invoke("deposit", args)
	//TODO: invoke chaincode
	//byteDeposit, err := chaincaller.Balance("deposit", string(byteBT))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	json.Unmarshal(byteDeposit, depo)

	now := currentLocalTime()
	//prepare  DepositRecord
	rec := models.DepositRecord{depo.DepositNo, depo.UserID, "add value to deposit", plus, depo.Balance, plus, "Deposit No.: " + depo.DepositNo, now}
	if err := models.AddDepositRecord(rec); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, depo)
}

func switchMeterInUse(c echo.Context) error {
	account := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["account"].(string)

	m := new(models.MeterDeposit)
	if err := c.Bind(m); err != nil {
		return err
	}

	meterInUse[account] = models.GetMeterDepositByID(m.MeterID)
	args := []string{meterInUse[account].DepositNo}
	byteDeposit, err := simulation.Invoke("query", args)
	//TODO: invoke chaincode get Balance
	//byteDeposit, err := chaincaller.Balance("query", meterInUse[account].DepositNo)
	if err != nil {
		return err
	}
	var depo models.Deposit
	json.Unmarshal(byteDeposit, &depo)

	buyKwh := models.GetUndealtKwhByMeter(meterInUse[account].MeterID, 1)
	sellKwh := models.GetUndealtKwhByMeter(meterInUse[account].MeterID, 2)
	saleable := models.GetLatestPowerSaleable(meterInUse[account].MeterID)
	stock := models.GetLatestPowerStocked(meterInUse[account].MeterID)
	//prd := models.GetLatestPowerRecord(meterInUse[account].MeterID)
	//return c.JSON(http.StatusOK, *meter)

	return c.JSON(http.StatusOK, map[string]string{
		"saleable":   fmt.Sprintf("%.2f", saleable.KwhSaleable),
		"stocked":    fmt.Sprintf("%.2f", stock.KwhStocked),
		"on_sell":    fmt.Sprintf("%.2f", sellKwh),
		"on_buy":     fmt.Sprintf("%.2f", buyKwh),
		"meter_id":   meterInUse[account].MeterID,
		"meter_name": meterInUse[account].MeterName,
		"deposit_no": meterInUse[account].DepositNo,
		"balance":    fmt.Sprintf("%.1f", depo.Balance)})
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

		meterInUse[u.Account] = models.GetDefaultMeterDeposit(u.ID)
		c.Response().Header().Set("Cache-Control", "no-cache")
		return c.JSON(http.StatusOK, map[string]string{
			"user_id":      strconv.Itoa(int(u.Model.ID)),
			"user_account": u.Account,
			"token":        t,
		})
	}

	return echo.ErrUnauthorized
}

func addPowerRecord(c echo.Context) error {
	type powerRec struct {
		KwhProduced float64
		KwhConsumed float64
		MeterID     string
	}
	pwr := new(powerRec)
	if err := c.Bind(pwr); err != nil {
		return err
	}

	if pwr.MeterID == "" {
		return c.String(http.StatusBadRequest, "MeterID can not be empty")
	}

	err := models.AddPowerRecord(pwr.KwhProduced, pwr.KwhConsumed, pwr.MeterID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, *pwr)
}

func currentLocalTime() time.Time {
	//local, _ := time.LoadLocation("Local")
	// local, _ := time.LoadLocation("Asia/Taipei")
	// now := time.Now().In(local)
	now := time.Now().UTC().Add(8 * time.Hour)
	return now
}
