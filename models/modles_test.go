package models

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

const checkMark = "\u2713"
const balloX = "\u2717"

func _TestAboutUser(t *testing.T) {

	t.Log("Begin unit tests of model User")
	//user := User{Account: "bruce.wu", Password: "0986", Attributes: []UserAttribute{}, PowerRecords: []PowerRecord{}}
	//
	//if err := AddUser(user); err != nil {
	//	t.Error(err, balloX)
	//}

	users, c := AllUsers(2, 2)
	fmt.Println("count: " + strconv.Itoa(c))
	fmt.Printf("%+v\n", users)
}

func _TestLoginCheck(t *testing.T) {
	t.Log("Begin unit tests of model User")

	_, logged := LoginCheck("gary.chang", "12345")
	if logged {
		t.Log("User login success.", checkMark)
	} else {
		t.Error("Account not exists", balloX)
	}
}

func _TestAddDepositRecord(t *testing.T) {
	now := time.Now().Local()
	fromRecord := DepositRecord{"test101", 42,
		"buy order expense", -30, 120,
		0, "Order Id: " + "oxoxoxox", now}
	AddDepositRecord(fromRecord)

}

func _TestAddPowerRecord(t *testing.T) {
	now := time.Now().Local()
	pr := PowerRecord{KwhProduced: 1000, KwhConsumed: -1000,
		KwhStocked: 0, KwhSaleable: 0, UpdatedAt: now, MeterID: "TM0042001"}

	prd, err := AddPowerRecord(&pr)
	if err != nil {
		t.Error(err, balloX)
	}

	fmt.Printf("%v+\n", prd)
}

func _TestAllPowerRecordsOfMeter(t *testing.T) {
	records, c := GetPowerRecordsByMeter("tm0002", 1, 10)
	//records, c := GetUserPowerRecords(2, 2, 2)
	fmt.Println("count: " + strconv.Itoa(c))
	for _, r := range records {
		fmt.Printf("Stock: %g, Pro: %g, Date: %v\n", r.KwhStocked, r.KwhProduced, r.UpdatedAt)
	}
}

func TestOrdersOfMeter(t *testing.T) {
	//orders, c := GetUndealtOrdersByMeter()

	//u := GetUndealtKwhByMeter("TM0042001", 1)
	//fmt.Printf("kwh: %f", u)
	// orders, c := GetUndealtOrdersByMeter("TM0044001", 1, 0, 10)
	// fmt.Println("count: " + strconv.Itoa(c))

	// orders := GetUndealtOrders(2)
	// for _, r := range orders {
	// 	fmt.Printf("%+v\n", r)
	// }

	err := UpdateOrderLocked("b8730005-2107-4532-b79f-345e64772e65", 1)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Println("-----------------------------------------------")
	// order2 := UpdateOrderLocked("55551f75-9e58-4601-9758-346f1b73c2fe", 88)
	// fmt.Printf("%+v\n", order2)
}

func _TestMeterDeposit(t *testing.T) {
	meter := GetDefaultMeterDeposit(3)
	//meter := GetLatestMeterDeposit(3)
	//meters := GetUserMeters(2)

	//for _, m := range meters {
	fmt.Printf("%+v\n", meter)
	//}
}
