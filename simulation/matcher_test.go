package simulation

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gitlab.com/wondervoyage/platform/models"
)

func _TestGoBuy(t *testing.T) {
	format := "2006-01-02 15:04:05"
	expire, _ := time.Parse(format, "2019-03-15 23:59:59")
	buy := models.Order{"bb001-1233-4421-a8eb-5144b9ca56d3",
		1,
		20,
		0,
		5,
		time.Now().Local(),
		expire,
		1,
		4,
		""}

	result, err := BuyHandler(buy)
	if err != nil {
		fmt.Println(err.Error())
	}
	var txns []models.DealTxn

	json.Unmarshal(result, &txns)

	for i, t := range txns {
		fmt.Printf("=====Txn %d ====: %+v\n", i, t)
	}

}

func _TestGoSell(t *testing.T) {

	format := "2006-01-02 15:04:05"
	expire, _ := time.Parse(format, "2019-03-15 23:59:59")
	sell := models.Order{"ss008-2534-4421-a8eb-5144b9ca56d3",
		2,
		15,
		0,
		5.3,
		time.Now().Local(),
		expire,
		1,
		2,
		""}

	/**********
			{  "Type": 1,
	  			"Kwh": 19,
	  			"Price": 3.8,
	  			"CreatedAt": "2019-03-16T12:30:00Z",
			    "ExpiredAt": "2019-03-16T12:30:00Z"
	  			"Status": 1,
	  			"UserID": 2
			}


	*************************/
	result, err := SellHandler(sell)
	if err != nil {
		fmt.Println(err.Error())
	}
	var txns []models.DealTxn

	json.Unmarshal(result, &txns)

	for i, t := range txns {
		fmt.Printf("=====Txn %d ====: %+v\n", i, t)
	}

}
