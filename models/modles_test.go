package models

import "testing"

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

func TestAddPowerRecord(t *testing.T) {
	pr := PowerRecord{KwhProduced: 10.5, KwhConsumed: 3.53, KwhSell: 0, KwhBuy: 0, UserID: 3}
	if err := AddPowerRecord(pr); err != nil {
		t.Error(err, balloX)
	}
}





