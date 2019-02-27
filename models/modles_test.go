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

func _TestAddPowerRecord(t *testing.T) {
	now := time.Now().Local()
	pr := PowerRecord{KwhProduced: 6.4, KwhConsumed: 2.3,
		KwhStocked: 0, KwhSaleable: 0, UpdatedAt: now, UserID: 3, MeterID: "pm0003"}

	if err := AddPowerRecord(pr); err != nil {
		t.Error(err, balloX)
	}

	//prd := GetLatestPowerRecord(2)
	//fmt.Printf("%v+\n", prd)
}

func _TestAllPowerRecordsOfUser(t *testing.T) {

	records, c := GetUserPowerRecords(2, 2, 2)
	fmt.Println("count: " + strconv.Itoa(c))
	for _, r := range records {
		fmt.Printf("Stock: %g, Pro: %g, Date: %v\n", r.KwhStocked, r.KwhProduced, r.UpdatedAt)
	}
}

func _TestOrdersOfUser(t *testing.T) {
	orders, c := GetUserUndealtOrders(2, 1, 1, 2)
	fmt.Println("count: " + strconv.Itoa(c))
	for _, r := range orders {
		fmt.Printf("%+v\n", r)
	}
}

func TestMeterDeposit(t *testing.T) {
	//meter := GetSelectedMeterDeposit(3)
	//meter := GetLatestMeterDeposit(3)
	meters := GetUserMeters(2)

	for _, m := range meters {
		fmt.Printf("%+v\n", m)
	}
}
