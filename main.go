package main

import (
	"fmt"

	"gitlab.com/wondervoyage/platform/chaincaller"
	_ "gitlab.com/wondervoyage/platform/chaincaller"
	"gitlab.com/wondervoyage/platform/models"
	_ "gitlab.com/wondervoyage/platform/models"
	"gitlab.com/wondervoyage/platform/rest"
	_ "gitlab.com/wondervoyage/platform/rest"
)

func main() {
	fmt.Println("Hello Platform !!")

	defer models.DB.Close()
	defer rest.RedisConn.Close()
	defer chaincaller.SDK.Close()

	rest.Router.Logger.Fatal(rest.Router.Start(":9080"))
}
