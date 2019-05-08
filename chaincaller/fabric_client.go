package chaincaller

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"gitlab.com/wondervoyage/platform/configs"
	"gitlab.com/wondervoyage/platform/models"
)

var SDK *fabsdk.FabricSDK

func init() {
	var err error
	//configProvider := config.FromFile("./config.yaml")
	SDK, err = fabsdk.New(config.FromFile(configs.Env.FabricPath))
	//SDK, err = fabsdk.New(config.FromFile("./configs/fabric/local/config.yaml"))
	//SDK, err = fabsdk.New(config.FromFile("./configs/fabric/staging/config.yaml"))
	//SDK, err = fabsdk.New(config.FromFile("/goapp/fabric/config.yaml"))
	if err != nil {
		log.Fatalf("create sdk fail: %s\n", err.Error())
	}
}

func PlaceOrder(action string, ord models.Order) ([]byte, error) {
	channelProvider := SDK.ChannelContext("mychannel", fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))
	channelClient, err := channel.New(channelProvider)
	if err != nil {
		log.Fatalf("create channel client fail: %s\n", err.Error())
	} else {
		fmt.Println("channelClient create successful !!!")
	}

	var args [][]byte
	jsonOrd, _ := json.Marshal(ord)
	args = append(args, jsonOrd)

	request := channel.Request{
		ChaincodeID: "order",
		Fcn:         action, // buy or sell
		Args:        args,
	}

	var response channel.Response
	var ex error

	if action == "query" {
		response, ex = channelClient.Query(request)
		if ex != nil {
			log.Fatal("query fail: ", ex.Error())
		} else {
			fmt.Printf("response is %s\n", response.Payload)
		}
	} else {
		response, ex = channelClient.Execute(request)
		if err != nil {
			log.Fatal("invoke fail: ", ex.Error())
		} else {
			fmt.Printf("response is %s\n", response.Payload)
		}
	}

	return response.Payload, ex
}

func Balance(action string, arg string) ([]byte, error) {
	channelProvider := SDK.ChannelContext("mychannel", fabsdk.WithUser("User1"), fabsdk.WithOrg("Org1"))
	channelClient, err := channel.New(channelProvider)
	if err != nil {
		log.Fatalf("create channel client fail: %s\n", err.Error())
	} else {
		fmt.Println("channelClient create successful !!!")
	}

	var args [][]byte
	args = append(args, []byte(arg))

	request := channel.Request{
		ChaincodeID: "balance",
		Fcn:         action, // create | transfer | deposit | withdraw | query
		Args:        args,
	}

	var response channel.Response
	var ex error

	if action == "query" {
		response, ex = channelClient.Query(request)
		if ex != nil {
			log.Fatal("query fail: ", ex.Error())
		} else {
			fmt.Printf("response is %s\n", response.Payload)
		}
	} else {
		response, ex = channelClient.Execute(request)
		if ex != nil {
			log.Fatal("invoke fail: ", err.Error())
		} else {
			fmt.Printf("response is %s\n", response.Payload)
		}
	}

	return response.Payload, err
}
