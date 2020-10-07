package main

import (
	"fmt"

	_ "gitlab.com/wondervoyage/platform/configs"
	"gitlab.com/wondervoyage/platform/models"
	_ "gitlab.com/wondervoyage/platform/models"
	"gitlab.com/wondervoyage/platform/rest"
	_ "gitlab.com/wondervoyage/platform/rest"
)

func main() {
	fmt.Println("Hello Platform !!")

	defer models.DB.Close()
	//defer chaincaller.SDK.Close()

	rest.Router.Logger.Fatal(rest.Router.Start(":9080"))
}
