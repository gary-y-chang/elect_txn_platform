package models

import (
	"testing"
	"time"
	"fmt"
)

const checkMark  = "\u2713"
const balloX  = "\u2717"

func _TestAddUser(t * testing.T) {

	t.Log("Begin unit tests of model User")
	user := User{Account: "bruce.wu", Password: "0986", Attributes: []UserAttribute{}, PowerRecords: []PowerRecord{}}


	if err := AddUser(user); err != nil {
		t.Error(err, balloX)
	}
}

func _TestLoginCheck(t *testing.T) {
	t.Log("Begin unit tests of model User")

	_, logged := LoginCheck("gary.chang", "12345")
	if logged{
		t.Log("User login success.", checkMark)
	}else {
		t.Error("Account not exists", balloX)
	}
}

func _TestAddPowerRecord(t *testing.T) {
	pr := PowerRecord{KwhProduced: 6.35, KwhConsumed: 4.00, KwhStocked: 8.62, UpdatedAt: time.Now(), UserID: 2}

	if err := AddPowerRecord(pr); err != nil {
		t.Error(err, balloX)
	}
}

func TestAllPowerRecordsOfUser(t *testing.T) {

	records := GetUserPowerRecords(2)

	for _, r := range records {
		fmt.Printf("Stock: %g, Pro: %g, Date: %v\n",r.KwhStocked, r.KwhProduced, r.UpdatedAt)
	}
}





