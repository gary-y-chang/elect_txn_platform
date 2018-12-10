package rest

import (
	"github.com/labstack/echo"
	"gitlab.com/wondervoyage/platform/models"
	"net/http"
	"github.com/dgrijalva/jwt-go"
	"time"
	"fmt"
)

func home(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome Home !")
}

func getAllUsers(c echo.Context) error {
    users := models.AllUsers()
	return c.JSON(http.StatusOK, users)
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
		return c.JSON(http.StatusOK, map[string]string{
			"token": t,
		})
	}

	return echo.ErrUnauthorized
}
