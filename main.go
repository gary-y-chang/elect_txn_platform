package main

import (
	"fmt"
	_ "gitlab.com/wondervoyage/platform/rest"
	"gitlab.com/wondervoyage/platform/rest"
)
func main() {
	fmt.Println("Hello Platform !!")
	rest.Router.Logger.Fatal(rest.Router.Start(":9080"))
}
