package main

import (
	"fmt"
	_ "gitlab.com/wondervoyage/platform/models"
	_ "gitlab.com/wondervoyage/platform/rest"
	"gitlab.com/wondervoyage/platform/rest"
	"gitlab.com/wondervoyage/platform/models"
)
func main() {
	fmt.Println("Hello Platform !!")
    defer models.DB.Close()

	rest.Router.Logger.Fatal(rest.Router.Start(":9080"))
}
