package simulation

import (
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"gitlab.com/wondervoyage/platform/models"
)

func BuyHandler(buy models.Order) ([]byte, error) {

	//level, err := leveldb.OpenFile("C:\\Database\\leveldb\\platform", nil)
	level, err := leveldb.OpenFile("/root/wondervoyage/leveldb", nil)
	if err != nil {
		panic(err)
	}
	defer level.Close()

	//  query Redis to check "Sell" list,  searching for the match selling order. If yes, update the status. if no, just push this buy order into Redis buying list
	//c, err := redis.Dial("tcp", "192.168.1.4:6379")
	// c, err := redis.Dial("tcp", "redis:6379")
	// if err != nil {
	// 	panic(err)
	// }
	// defer c.Close()

	//txnId := uuid.Must(uuid.NewV4()).String()
	txnId := uuid.NewV4().String()
	txns := make([]models.DealTxn, 0)

	//st, err := redis.ByteSlices(c.Do("LRANGE", "simu-sell-list", 0, -1))
	var match bool = true
	for match {
		sales := models.GetUndealtOrders(2) //sell-list
		for i, sale := range sales {
			//fmt.Printf("index %d    %#v\t%s\t%s\t%f\n", i, s, s.Id, s.Name, s.Height)
			if buy.MeterID == sale.MeterID {
				continue
			}
			if buy.Price >= sale.Price {
				//order match !!!!
				fmt.Printf("Matching sell order : index %d\t%#v\n", i, sale)
				remains := sale.Kwh - sale.KwhDealt
				unfilled := buy.Kwh - buy.KwhDealt
				if remains-unfilled < 0 { // buy not-fulfilled, sold out
					txn := models.DealTxn{}
					txn.Kwh = remains

					// remove this sale from sell-list
					err := models.UpdateOrderLocked(sale.ID, txn.Kwh)
					if err != nil {
						fmt.Println(err.Error())
						break
					}

					buy.KwhDealt = buy.KwhDealt + txn.Kwh
					// then, update the state in the ledger
					sale.KwhDealt = sale.KwhDealt + txn.Kwh
					sale.Status = 2
					jsonSale, err := json.Marshal(sale)
					err = level.Put([]byte(sale.ID), jsonSale, nil)
					//making one txn
					txn.ID = txnId
					fmt.Printf("The TxnID: %s\n", txn.ID)
					//txn.Kwh =
					txn.Price = sale.Price
					txn.TxnDate = time.Now().UTC().Add(8 * time.Hour)
					txn.BuyOrderID = buy.ID
					txn.BuyDepositNo = buy.DepositNo
					txn.SellOrderID = sale.ID
					txn.SellDepositNo = sale.DepositNo
					txns = append(txns, txn)

					continue
				} else if remains-unfilled > 0 { // buy fulfilled, not sold out
					txn := models.DealTxn{}
					txn.Kwh = unfilled

					err := models.UpdateOrderLocked(sale.ID, txn.Kwh)
					if err != nil {
						fmt.Println(err.Error())
						break
					}

					sale.KwhDealt = sale.KwhDealt + txn.Kwh
					//update this sell order to sell-list
					jsonSale, err := json.Marshal(sale)
					err = level.Put([]byte(sale.ID), jsonSale, nil)

					buy.KwhDealt = buy.KwhDealt + txn.Kwh
					buy.Status = 2
					jsonBuy, err := json.Marshal(buy)
					err = level.Put([]byte(buy.ID), jsonBuy, nil)

					txn.ID = txnId
					fmt.Printf("The TxnID: %s\n", txn.ID)
					//txn.Kwh =
					txn.Price = sale.Price
					txn.TxnDate = time.Now().UTC().Add(8 * time.Hour)
					txn.BuyOrderID = buy.ID
					txn.BuyDepositNo = buy.DepositNo
					txn.SellOrderID = sale.ID
					txn.SellDepositNo = sale.DepositNo
					txns = append(txns, txn)
					jsonTxn, err := json.Marshal(txns)

					return jsonTxn, err
				} else if remains-unfilled == 0 { // buy fulfilled, sold out
					txn := models.DealTxn{}
					txn.Kwh = remains
					err := models.UpdateOrderLocked(sale.ID, txn.Kwh)
					if err != nil {
						fmt.Println(err.Error())
						break
					}

					// update the state in the ledger
					sale.KwhDealt = sale.KwhDealt + txn.Kwh
					sale.Status = 2
					//update this sell order to sell-list
					jsonSale, err := json.Marshal(sale)
					err = level.Put([]byte(sale.ID), jsonSale, nil)

					buy.KwhDealt = buy.KwhDealt + txn.Kwh
					buy.Status = 2
					jsonBuy, err := json.Marshal(buy)
					err = level.Put([]byte(buy.ID), jsonBuy, nil)

					txn.ID = txnId
					fmt.Printf("The TxnID: %s\n", txn.ID)
					//txn.Kwh =
					txn.Price = sale.Price
					txn.TxnDate = time.Now().UTC().Add(8 * time.Hour)
					txn.BuyOrderID = buy.ID
					txn.BuyDepositNo = buy.DepositNo
					txn.SellOrderID = sale.ID
					txn.SellDepositNo = sale.DepositNo
					txns = append(txns, txn)
					jsonTxn, err := json.Marshal(txns)

					return jsonTxn, err
				}
			}
		}
		match = false
	}

	ordBoard := new(models.OrderBoard)
	models.AddOrderBoard(ordBoard.Transmit(buy))

	jsonOrder, err := json.Marshal(buy)
	err = level.Put([]byte(buy.ID), jsonOrder, nil)

	txn := models.DealTxn{}
	txns = append(txns, txn)
	jsonTxn, err := json.Marshal(txns) // this will return an empty txn

	return jsonTxn, err
}

func SellHandler(sell models.Order) ([]byte, error) {

	//level, err := leveldb.OpenFile("C:\\Database\\leveldb\\platform", nil)
	level, err := leveldb.OpenFile("/root/wondervoyage/leveldb", nil)
	if err != nil {
		panic(err)
	}
	defer level.Close()

	/*** Logic:
	     Query to check "Buy" list,  searching for the match selling order.
	     If yes, update the status. if no, just push this buy order into the Redis buying list
	***/

	//txnId := uuid.Must(uuid.NewV4()).String()
	txnId := uuid.NewV4().String()
	txns := make([]models.DealTxn, 0)
	//st, err := redis.ByteSlices(c.Do("LRANGE", "simu-buy-list", 0, -1))
	var match bool = true
	for match {
		buyes := models.GetUndealtOrders(1) //buy-list
		for i, buy := range buyes {
			if sell.MeterID == buy.MeterID {
				continue
			}
			if buy.Price >= sell.Price {
				//price match , deal!!
				fmt.Printf("Matching buying order : index %d\t%#v\n", i, buy)
				// Check the quantities for sale and update the remaining quantities.
				// If sold out, do LPOP to remove the most left element; if not sold out, update the remaining quantites of the sale order

				remains := buy.Kwh - buy.KwhDealt
				unfilled := sell.Kwh - sell.KwhDealt
				if remains-unfilled < 0 { //buy fulfilled, not sold out

					txn := models.DealTxn{}
					txn.Kwh = remains
					err := models.UpdateOrderLocked(buy.ID, txn.Kwh)
					if err != nil {
						fmt.Println(err.Error())
						break
					}

					//then, update the state in the ledger
					buy.KwhDealt = buy.KwhDealt + txn.Kwh
					buy.Status = 2
					jsonBuy, err := json.Marshal(buy)
					//err = stub.PutState(buy.ID, jsonBuy)
					err = level.Put([]byte(buy.ID), jsonBuy, nil)

					sell.KwhDealt = sell.KwhDealt + txn.Kwh
					//making one txn
					txn.ID = txnId
					fmt.Printf("The TxnID: %s\n", txn.ID)
					//txn.Kwh =
					txn.Price = sell.Price
					txn.TxnDate = time.Now().UTC().Add(8 * time.Hour)
					txn.BuyOrderID = buy.ID
					txn.BuyDepositNo = buy.DepositNo
					txn.SellOrderID = sell.ID
					txn.SellDepositNo = sell.DepositNo
					txns = append(txns, txn)

					continue
				} else if remains-unfilled > 0 { //buy not-fulfilled, sold out
					txn := models.DealTxn{}
					txn.Kwh = unfilled
					err := models.UpdateOrderLocked(buy.ID, txn.Kwh)
					if err != nil {
						fmt.Println(err.Error())
						break
					}

					buy.KwhDealt = buy.KwhDealt + txn.Kwh
					jsonBuy, err := json.Marshal(buy)
					err = level.Put([]byte(buy.ID), jsonBuy, nil)

					sell.KwhDealt = sell.KwhDealt + txn.Kwh
					sell.Status = 2
					jsonOdr, err := json.Marshal(sell)
					err = level.Put([]byte(sell.ID), jsonOdr, nil)

					//txn.ID = stub.GetTxID()
					txn.ID = txnId
					fmt.Printf("The TxnID: %s\n", txn.ID)
					//txn.Kwh =
					txn.Price = sell.Price
					txn.TxnDate = time.Now().UTC().Add(8 * time.Hour)
					txn.BuyOrderID = buy.ID
					txn.BuyDepositNo = buy.DepositNo
					txn.SellOrderID = sell.ID
					txn.SellDepositNo = sell.DepositNo
					txns = append(txns, txn)
					jsonTxn, err := json.Marshal(txns)

					return jsonTxn, err
				} else if remains-unfilled == 0 { //buy fulfilled, sold out
					txn := models.DealTxn{}
					txn.Kwh = remains
					err := models.UpdateOrderLocked(buy.ID, txn.Kwh)
					if err != nil {
						fmt.Println(err.Error())
						break
					}

					//then, update the state in the ledger
					buy.KwhDealt = buy.KwhDealt + txn.Kwh
					buy.Status = 2
					jsonBuy, err := json.Marshal(buy)
					err = level.Put([]byte(buy.ID), jsonBuy, nil)

					sell.KwhDealt = sell.KwhDealt + txn.Kwh
					sell.Status = 2
					jsonOdr, err := json.Marshal(sell)
					err = level.Put([]byte(sell.ID), jsonOdr, nil)

					txn.ID = txnId
					fmt.Printf("The TxnID: %s\n", txn.ID)
					//txn.Kwh =
					txn.Price = sell.Price
					txn.TxnDate = time.Now().UTC().Add(8 * time.Hour)
					txn.BuyOrderID = buy.ID
					txn.BuyDepositNo = buy.DepositNo
					txn.SellOrderID = sell.ID
					txn.SellDepositNo = sell.DepositNo
					txns = append(txns, txn)
					jsonTxn, err := json.Marshal(txns)

					return jsonTxn, err
				}
			}
		}
		match = false
	}

	ordBoard := new(models.OrderBoard)
	models.AddOrderBoard(ordBoard.Transmit(sell))

	jsonOrder, err := json.Marshal(sell)
	err = level.Put([]byte(sell.ID), jsonOrder, nil)
	txn := models.DealTxn{}
	txns = append(txns, txn)
	jsonTxn, err := json.Marshal(txns) // this will return an empty txn

	return jsonTxn, err
}
