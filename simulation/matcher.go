package simulation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"gitlab.com/wondervoyage/platform/models"
)

func BuyHandler(buy models.Order) ([]byte, error) {

	level, err := leveldb.OpenFile("C:\\Database\\leveldb\\platform", nil)
	//level, err := leveldb.OpenFile("/root/wondervoyage/leveldb", nil)
	if err != nil {
		panic(err)
	}
	defer level.Close()

	//  query Redis to check "Sell" list,  searching for the match selling order. If yes, update the status. if no, just push this buy order into Redis buying list
	c, err := redis.Dial("tcp", "192.168.1.4:6379")
	//c, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	//txnId := uuid.Must(uuid.NewV4()).String()
	txnId := uuid.NewV4().String()
	txns := make([]models.DealTxn, 0)
	st, err := redis.ByteSlices(c.Do("LRANGE", "simu-sell-list", 0, -1))
	for i, o := range st {
		var sale models.Order
		err = json.Unmarshal(o, &sale)
		//fmt.Printf("index %d    %#v\t%s\t%s\t%f\n", i, s, s.Id, s.Name, s.Height)
		if buy.UserID == sale.UserID {
			continue
		}
		if buy.Price >= sale.Price {
			//order match !!!!
			fmt.Printf("Matching sell order : index %d\t%#v\n", i, sale)
			remains := sale.Kwh - sale.KwhDealt
			unfilled := buy.Kwh - buy.KwhDealt
			if remains-unfilled < 0 { // buy not-fulfilled, sold out
				txn := models.DealTxn{}
				// remove this sale from sell-list
				v, err := c.Do("LREM", "simu-sell-list", 0, o)
				if err != nil {
					fmt.Printf(err.Error())
				} else {
					fmt.Printf("Redis Reply: %#v\n", v)
				}
				// then, update the state in the ledger
				txn.Kwh = remains
				buy.KwhDealt = buy.KwhDealt + txn.Kwh
				//jsonBuy, err := json.Marshal(*buy)
				//err = level.Put([]byte(buy.ID), jsonBuy, nil)
				//v, err = c.Do("RPUSH", "simu-buy-list", jsonBuy)
				//if err != nil {
				//	fmt.Printf(err.Error())
				//}else {
				//	fmt.Printf("Redis Reply: %#v\n", v)
				//}

				sale.KwhDealt = sale.KwhDealt + txn.Kwh
				sale.Status = 2
				jsonSale, err := json.Marshal(sale)
				err = level.Put([]byte(sale.ID), jsonSale, nil)
				//making one txn
				txn.ID = txnId
				fmt.Printf("The TxnID: %s\n", txn.ID)
				//txn.Kwh =
				txn.Price = sale.Price
				local, _ := time.LoadLocation("Local")
				txn.TxnDate = time.Now().In(local)
				txn.BuyOrderID = buy.ID
				txn.BuyDepositNo = buy.DepositNo
				txn.SellOrderID = sale.ID
				txn.SellDepositNo = sale.DepositNo
				txns = append(txns, txn)

				continue
			} else if remains-unfilled > 0 { // buy fulfilled, not sold out
				txn := models.DealTxn{}
				txn.Kwh = unfilled
				sale.KwhDealt = sale.KwhDealt + txn.Kwh
				//update this sell order to sell-list
				jsonSale, err := json.Marshal(sale)
				err = level.Put([]byte(sale.ID), jsonSale, nil)
				//v, err := c.Do("LSET", "simu-sell-list", i, jsonSale)

				// remove the original sale from sell-list
				v, err := c.Do("LREM", "simu-sell-list", 0, o)
				// push the updated sale from the right
				v, err = c.Do("RPUSH", "simu-sell-list", jsonSale)
				if err != nil {
					fmt.Printf(err.Error())
				} else {
					fmt.Printf("Redis Reply: %#v\n", v)
				}

				buy.KwhDealt = buy.KwhDealt + txn.Kwh
				buy.Status = 2
				jsonBuy, err := json.Marshal(buy)
				err = level.Put([]byte(buy.ID), jsonBuy, nil)

				txn.ID = txnId
				fmt.Printf("The TxnID: %s\n", txn.ID)
				//txn.Kwh =
				txn.Price = sale.Price
				local, _ := time.LoadLocation("Local")
				txn.TxnDate = time.Now().In(local)
				txn.BuyOrderID = buy.ID
				txn.BuyDepositNo = buy.DepositNo
				txn.SellOrderID = sale.ID
				txn.SellDepositNo = sale.DepositNo
				txns = append(txns, txn)
				jsonTxn, err := json.Marshal(txns)

				return jsonTxn, err
			} else if remains-unfilled == 0 { // buy fulfilled, sold out
				txn := models.DealTxn{}
				// remove this sale from sell-list
				v, err := c.Do("LREM", "simu-sell-list", 0, o)
				if err != nil {
					fmt.Printf(err.Error())
				} else {
					fmt.Printf("Redis Reply: %#v\n", v)
				}
				// then, update the state in the ledger
				txn.Kwh = remains
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
				local, _ := time.LoadLocation("Local")
				txn.TxnDate = time.Now().In(local)
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

	jsonOrder, err := json.Marshal(buy)
	v, err := c.Do("RPUSH", "simu-buy-list", jsonOrder)
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		fmt.Printf("Redis Reply: %#v\n", v)
	}
	err = level.Put([]byte(buy.ID), jsonOrder, nil)
	txn := models.DealTxn{}
	txns = append(txns, txn)
	jsonTxn, err := json.Marshal(txns) // this will return an empty txn

	return jsonTxn, err
}

func SellHandler(sell models.Order) ([]byte, error) {

	level, err := leveldb.OpenFile("C:\\Database\\leveldb\\platform", nil)
	//level, err := leveldb.OpenFile("/root/wondervoyage/leveldb", nil)
	if err != nil {
		panic(err)
	}
	defer level.Close()

	/*** Logic:
	     Query the Redis to check "Buy" list,  searching for the match selling order.
	     If yes, update the status. if no, just push this buy order into the Redis buying list
	***/
	c, err := redis.Dial("tcp", "192.168.1.4:6379")
	//c, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	//txnId := uuid.Must(uuid.NewV4()).String()
	txnId := uuid.NewV4().String()
	txns := make([]models.DealTxn, 0)
	st, err := redis.ByteSlices(c.Do("LRANGE", "simu-buy-list", 0, -1))
	for i, o := range st {
		var buy models.Order
		err = json.Unmarshal(o, &buy)
		//fmt.Printf("index %d    %#v\t%s\t%s\t%f\n", i, s, s.Id, s.Name, s.Height)
		if sell.UserID == buy.UserID {
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
				//remove this buy order from buy-list
				v, err := c.Do("LREM", "simu-buy-list", 0, o)
				if err != nil {
					fmt.Printf(err.Error())
				} else {
					fmt.Printf("Redis Reply: %#v\n", v)
				}
				//then, update the state in the ledger
				txn := models.DealTxn{}
				txn.Kwh = remains
				buy.KwhDealt = buy.KwhDealt + txn.Kwh
				buy.Status = 2
				jsonBuy, err := json.Marshal(buy)
				//err = stub.PutState(buy.ID, jsonBuy)
				err = level.Put([]byte(buy.ID), jsonBuy, nil)

				//jsonOdr, err := json.Marshal(*sell)
				//err = level.Put([]byte(sell.ID), jsonOdr, nil)
				//v, err = c.Do("RPUSH", "simu-sell-list", jsonOdr)
				//if err != nil {
				//	fmt.Printf(err.Error())
				//}else {
				//	fmt.Printf("Redis Reply: %#v\n", v)
				//}

				sell.KwhDealt = sell.KwhDealt + txn.Kwh
				//making one txn
				txn.ID = txnId
				fmt.Printf("The TxnID: %s\n", txn.ID)
				//txn.Kwh =
				txn.Price = sell.Price
				local, _ := time.LoadLocation("Local")
				txn.TxnDate = time.Now().In(local)
				txn.BuyOrderID = buy.ID
				txn.BuyDepositNo = buy.DepositNo
				txn.SellOrderID = sell.ID
				txn.SellDepositNo = sell.DepositNo
				txns = append(txns, txn)

				continue
			} else if remains-unfilled > 0 { //buy not-fulfilled, sold out
				txn := models.DealTxn{}
				txn.Kwh = unfilled
				buy.KwhDealt = buy.KwhDealt + txn.Kwh
				//update this buy order to buy-list
				jsonBuy, err := json.Marshal(buy)
				err = level.Put([]byte(buy.ID), jsonBuy, nil)
				//v, err := c.Do("LSET", "simu-buy-list", i, jsonBuy)

				//remove the original buy order from buy-list
				v, err := c.Do("LREM", "simu-buy-list", 0, o)
				//push the updated buy order to the right of buy-list
				v, err = c.Do("RPUSH", "simu-buy-list", jsonBuy)
				if err != nil {
					fmt.Printf(err.Error())
				} else {
					fmt.Printf("Redis Reply: %#v\n", v)
				}

				sell.KwhDealt = sell.KwhDealt + txn.Kwh
				sell.Status = 2
				jsonOdr, err := json.Marshal(sell)
				err = level.Put([]byte(sell.ID), jsonOdr, nil)

				//txn.ID = stub.GetTxID()
				txn.ID = txnId
				fmt.Printf("The TxnID: %s\n", txn.ID)
				//txn.Kwh =
				txn.Price = sell.Price
				txn.TxnDate = time.Now().Local()
				txn.BuyOrderID = buy.ID
				txn.BuyDepositNo = buy.DepositNo
				txn.SellOrderID = sell.ID
				txn.SellDepositNo = sell.DepositNo
				txns = append(txns, txn)
				jsonTxn, err := json.Marshal(txns)

				return jsonTxn, err
			} else if remains-unfilled == 0 { //buy fulfilled, sold out
				txn := models.DealTxn{}
				//remove this buy order from buy-list
				v, err := c.Do("LREM", "simu-buy-list", 0, o)
				if err != nil {
					fmt.Printf(err.Error())
				} else {
					fmt.Printf("Redis Reply: %#v\n", v)
				}
				//then, update the state in the ledger
				txn.Kwh = remains
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
				local, _ := time.LoadLocation("Local")
				txn.TxnDate = time.Now().In(local)
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

	jsonOrder, err := json.Marshal(sell)
	v, err := c.Do("RPUSH", "simu-sell-list", jsonOrder)
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		fmt.Printf("Redis Reply: %#v\n", v)
	}
	err = level.Put([]byte(sell.ID), jsonOrder, nil)
	txn := models.DealTxn{}
	txns = append(txns, txn)
	jsonTxn, err := json.Marshal(txns) // this will return an empty txn

	return jsonTxn, err
}
