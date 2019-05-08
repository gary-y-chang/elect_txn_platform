package simulation

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"gitlab.com/wondervoyage/platform/configs"
)

type Deposit struct {
	DepositNo string
	Balance   float64
	Payable   float64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uint
}

type BalanceTxn struct {
	Target string  // the deposit_no to input
	Source string  // the deposit_no to output
	Amount float64 // the $$
}

// return the Deposit involved. If two involved, [From_Deposit, To_Deposit]
func Invoke(fn string, args []string) ([]byte, error) {

	level, err := leveldb.OpenFile(configs.Env.LeveldbPath, nil)
	//level, err := leveldb.OpenFile("/root/wondervoyage/leveldb", nil)
	if err != nil {
		panic(err)
	}
	defer level.Close()

	var txn BalanceTxn
	var depos Deposit
	if fn == "query" {
		fmt.Printf("Query for deposit no.: %s\n", args[0])
	} else if fn == "create" {
		byteDepo := []byte(args[0])

		json.Unmarshal(byteDepo, &depos)
		fmt.Printf("Create new deposit %+v\n", depos)
	} else {
		byteTxn := []byte(args[0])

		json.Unmarshal(byteTxn, &txn)
		fmt.Printf("Txn unmarshaled: %+v\n", txn)
	}

	var result string

	switch fn {
	case "create":
		result, err = create(level, &depos)

	case "detain":
		result, err = detain(level, &txn)

	case "transfer":
		result, err = transfer(level, &txn)

	case "payOrder":
		result, err = payOrder(level, &txn)

	case "deposit":
		result, err = deposit(level, &txn)

	case "withdraw":
		result, err = withdraw(level, &txn)

	case "query":
		result, err = query(level, args[0])

	default:
		err = errors.New("function invocation error")
	}

	//if err != nil {
	//	return shim.Error(err.Error())
	//}

	// Return the result as success payload
	return []byte(result), err
}

func create(level *leveldb.DB, depos *Deposit) (string, error) {
	//local, _ := time.LoadLocation("Local")
	//now := time.Now().In(local)
	//userId, _ := strconv.ParseUint(uid, 10, 64)
	//deposit := models.Deposit{depoNo,
	//0,0,now,now,userId}

	jsonDepo, _ := json.Marshal(depos)
	level.Put([]byte(depos.DepositNo), jsonDepo, nil)

	return string(jsonDepo), nil
}

func detain(level *leveldb.DB, txn *BalanceTxn) (string, error) {

	//byteDeposit, err := stub.GetState(txn.Target)
	byteDeposit, err := level.Get([]byte(txn.Target), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + txn.Target + "\"}", err
	} else if byteDeposit == nil {
		return "{\"Error\":\"Deposit: " + txn.Target + " does not exist.\"}", err
	}
	var depo Deposit
	json.Unmarshal(byteDeposit, &depo)

	depo.Balance = depo.Balance - txn.Amount
	depo.Payable = depo.Payable + txn.Amount
	local, _ := time.LoadLocation("Local")
	depo.UpdatedAt = time.Now().In(local)
	jsonDepo, err := json.Marshal(depo)
	//err = stub.PutState(depo.DepositNo, jsonDepo)
	err = level.Put([]byte(depo.DepositNo), jsonDepo, nil)
	if err != nil {
		return "{\"Error\":\"Wrong payable update of Deposit: " + txn.Target + "\"}", err
	}
	return string(jsonDepo), nil
}

func transfer(level *leveldb.DB, txn *BalanceTxn) (string, error) {
	//byteTarget, err :=stub.GetState(txn.Target)
	byteTarget, err := level.Get([]byte(txn.Target), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + txn.Target + "\"}", err
	} else if byteTarget == nil {
		return "{\"Error\":\"Deposit: " + txn.Target + " does not exist.\"}", err
	}
	var depoTarget Deposit
	json.Unmarshal(byteTarget, &depoTarget)

	//byteSource, err :=stub.GetState(txn.Source)
	byteSource, err := level.Get([]byte(txn.Source), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + txn.Source + "\"}", err
	} else if byteTarget == nil {
		return "{\"Error\":\"Deposit: " + txn.Source + " does not exist.\"}", err
	}
	var depoSource Deposit
	json.Unmarshal(byteSource, &depoSource)

	local, _ := time.LoadLocation("Local")
	depoSource.Balance = depoSource.Balance - txn.Amount
	depoSource.UpdatedAt = time.Now().In(local)

	depoTarget.Balance = depoTarget.Balance + txn.Amount
	depoTarget.UpdatedAt = time.Now().In(local)

	jsonDepoTarget, err := json.Marshal(depoTarget)
	//err = stub.PutState(depoTarget.DepositNo, jsonDepoTarget)
	err = level.Put([]byte(depoTarget.DepositNo), jsonDepoTarget, nil)
	if err != nil {
		return "{\"Error\":\"Wrong balance update of Deposit: " + txn.Target + "\"}", err
	}

	jsonDepoSource, err := json.Marshal(depoSource)
	//err = stub.PutState(depoSource.DepositNo, jsonDepoSource)
	err = level.Put([]byte(depoSource.DepositNo), jsonDepoSource, nil)
	if err != nil {
		return "{\"Error\":\"Wrong balance update of Deposit: " + txn.Source + "\"}", err
	}

	result := make([]Deposit, 0)
	result = append(result, depoSource, depoTarget)
	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

func payOrder(level *leveldb.DB, txn *BalanceTxn) (string, error) {
	//byteTarget, err :=stub.GetState(txn.Target)
	byteTarget, err := level.Get([]byte(txn.Target), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + txn.Target + "\"}", err
	} else if byteTarget == nil {
		return "{\"Error\":\"Deposit: " + txn.Target + " does not exist.\"}", err
	}
	var depoTarget Deposit
	json.Unmarshal(byteTarget, &depoTarget)

	//byteSource, err :=stub.GetState(txn.Source)
	byteSource, err := level.Get([]byte(txn.Source), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + txn.Source + "\"}", err
	} else if byteTarget == nil {
		return "{\"Error\":\"Deposit: " + txn.Source + " does not exist.\"}", err
	}
	var depoSource Deposit
	json.Unmarshal(byteSource, &depoSource)

	local, _ := time.LoadLocation("Local")
	depoSource.Payable = depoSource.Payable - txn.Amount
	depoSource.UpdatedAt = time.Now().In(local)

	depoTarget.Balance = depoTarget.Balance + txn.Amount
	depoTarget.UpdatedAt = time.Now().In(local)

	jsonDepoTarget, err := json.Marshal(depoTarget)
	//err = stub.PutState(depoTarget.DepositNo, jsonDepoTarget)
	err = level.Put([]byte(depoTarget.DepositNo), jsonDepoTarget, nil)
	if err != nil {
		return "{\"Error\":\"Wrong balance update of Deposit: " + txn.Target + "\"}", err
	}

	jsonDepoSource, err := json.Marshal(depoSource)
	//err = stub.PutState(depoSource.DepositNo, jsonDepoSource)
	err = level.Put([]byte(depoSource.DepositNo), jsonDepoSource, nil)
	if err != nil {
		return "{\"Error\":\"Wrong balance update of Deposit: " + txn.Source + "\"}", err
	}

	result := make([]Deposit, 0)
	result = append(result, depoSource, depoTarget)
	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

func deposit(level *leveldb.DB, txn *BalanceTxn) (string, error) {
	//byteTarget, err :=stub.GetState(txn.Target)
	byteTarget, err := level.Get([]byte(txn.Target), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + txn.Target + "\"}", err
	} else if byteTarget == nil {
		return "{\"Error\":\"Deposit: " + txn.Target + " does not exist.\"}", err
	}
	var depo Deposit
	json.Unmarshal(byteTarget, &depo)

	depo.Balance = depo.Balance + txn.Amount

	local, _ := time.LoadLocation("Local")
	depo.UpdatedAt = time.Now().In(local)
	jsonDepo, err := json.Marshal(depo)
	//err = stub.PutState(depo.DepositNo, jsonDepo)
	err = level.Put([]byte(depo.DepositNo), jsonDepo, nil)
	if err != nil {
		return "{\"Error\":\"Wrong adding balance of Deposit: " + txn.Target + "\"}", err
	}
	return string(jsonDepo), nil

}

func withdraw(level *leveldb.DB, txn *BalanceTxn) (string, error) {
	//byteTarget, err := stub.GetState(txn.Target)
	byteTarget, err := level.Get([]byte(txn.Target), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + txn.Target + "\"}", err
	} else if byteTarget == nil {
		return "{\"Error\":\"Deposit: " + txn.Target + " does not exist.\"}", err
	}
	var depo Deposit
	json.Unmarshal(byteTarget, &depo)

	depo.Balance = depo.Balance - txn.Amount

	local, _ := time.LoadLocation("Local")
	depo.UpdatedAt = time.Now().In(local)
	jsonDepo, err := json.Marshal(depo)
	//err = stub.PutState(depo.DepositNo, jsonDepo)
	err = level.Put([]byte(depo.DepositNo), jsonDepo, nil)
	if err != nil {
		return "{\"Error\":\"Wrong withdrawing balance of Deposit: " + txn.Target + "\"}", err
	}
	return string(jsonDepo), nil
}

func query(level *leveldb.DB, depoNo string) (string, error) {
	//byteDepo, err :=stub.GetState(depoNo)
	byteDepo, err := level.Get([]byte(depoNo), nil)
	if err != nil {
		return "{\"Error\":\"Failed to get Deposit state for " + depoNo + "\"}", err
	} else if byteDepo == nil {
		return "{\"Error\":\"Deposit: " + depoNo + " does not exist.\"}", err
	}

	return string(byteDepo), nil
}
