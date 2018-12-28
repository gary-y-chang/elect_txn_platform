package rest

import (
	"github.com/labstack/echo"
	"gitlab.com/wondervoyage/platform/models"
	"net/http"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
	"strconv"
)

func home(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome Home !")
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
