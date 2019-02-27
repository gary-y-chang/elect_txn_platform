package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type User struct {
	gorm.Model
	Account      string `gorm:"unique;not null"`
	Password     string `gorm:"not null"`
	Name         string
	Attributes   []UserAttribute
	PowerRecords []PowerRecord
}

type UserAttribute struct {
	UserID  uint
	Email   string
	Address string
	Phone   string
}

type MeterDeposit struct {
	MeterID     string `gorm:"primary_key;not null"` // the smart meter ID
	MeterName   string
	DepositNo   string `gorm:"unique;not null"`
	BankAccount string // binding a bank account
	IsDefault   bool   // by default,  the user logged-in to be used
	UserID      uint
	OrgID       uint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PowerRecord struct {
	KwhProduced float64 // power kwh produced and saved to 儲能櫃 during a time period, not accumulated
	KwhConsumed float64 // power kwh consumed from 儲能櫃 during a time period, not accumulated
	KwhStocked  float64 // power kwh accumulated from buy-order, stocked = stocked - consumed
	KwhSaleable float64 // power kwh accumulated from produced, saleable = saleable + produced
	UpdatedAt   time.Time
	MeterID     string
	UserID      uint
}

/*** e.g
{ "kwh": 25.5,
  "price": 6.8,
  "created": "2019-01-15T04:00:00.000Z",
  "expired": "2019-02-15T04:00:00.000Z",
  "user-id": 2
}
 **/
type Order struct {
	ID        string  // UUID           timestamp := time.Now().Unix()
	Type      uint8   // 1:buy, 2:sell
	Kwh       float64 // Kwh to be sold
	KwhDealt  float64 // Kwh already sold
	Price     float64
	CreatedAt time.Time
	ExpiredAt time.Time
	Status    uint8 //1:un-dealt, 2:dealt, 3:expired, 4:canceled
	UserID    uint
	DepositNo string
}

type DealTxn struct {
	ID            string `gorm:"primary_key"` // transaction id from chaincode
	Part          uint8  `gorm:"primary_key"`
	Kwh           float64
	Price         float64   // dealt price
	TxnDate       time.Time // display format 2018-11-20 23:45
	BuyOrderID    string    // order id
	BuyDepositNo  string
	SellOrderID   string // order id
	SellDepositNo string
}

type DepositRecord struct {
	DepositNo string
	UserID    uint
	Action    string
	InOut     float64 // in with positive number, out with negative number
	Balance   float64 // remained balance
	Money     int
	Memo      string
	RecDate   time.Time
}

//  for use on chaincode
type Deposit struct {
	DepositNo string
	Balance   float64
	Payable   float64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uint
}

// for use on chaincode
type BalanceTxn struct {
	Target string  // the deposit_no to input
	Source string  // the deposit_no to output
	Amount float64 // the $$
}

var DB *gorm.DB

func init() {
	var err error
	//DB, err = gorm.Open("postgres", "host=192.168.43.214 port=15432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	DB, err = gorm.Open("postgres", "host=192.168.1.4 port=15432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	//DB, err = gorm.Open("postgres", "host=pgdb port=5432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	DB.LogMode(true)
	DB.SingularTable(true)
	DB.AutoMigrate(&User{}, &UserAttribute{}, &PowerRecord{}, &Order{}, &DealTxn{}, &DepositRecord{}, &MeterDeposit{})
}

func AddUser(u User) error {
	if DB.Create(&u).Error != nil {
		return errors.New("user already exists")
	}
	return nil
}

func AddMeterDeposit(md MeterDeposit) error {
	if DB.Create(&md).Error != nil {
		return errors.New("MeterDeposit already exists")
	}
	return nil
}

func AddOrder(odr Order) error {
	if DB.Create(&odr).Error != nil {
		return errors.New("error while adding Order")
	}
	return nil
}

func AddDealTxn(txn DealTxn) error {
	if DB.Create(&txn).Error != nil {
		return errors.New("error while adding DealTxn")
	}
	return nil
}

func AddDepositRecord(drecord DepositRecord) error {
	if DB.Create(&drecord).Error != nil {
		return errors.New("error while adding DepositRecord")
	}
	return nil
}

func AddPowerRecord(pr PowerRecord) error {
	lastPR := GetLatestPowerRecord(pr.UserID)
	pr.KwhStocked = lastPR.KwhStocked - pr.KwhConsumed
	pr.KwhSaleable = lastPR.KwhSaleable + pr.KwhProduced
	if DB.Create(&pr).Error != nil {
		return errors.New("error while adding PowerRecord")
	}
	return nil
}

func UpdatePowerRecordByTxn(txn DealTxn) error {
	local, _ := time.LoadLocation("Local")
	now := time.Now().In(local)

	var buy, sell Order
	DB.First(&buy, "id = ?", txn.BuyOrderID)
	DB.First(&sell, "id = ?", txn.SellOrderID)

	brecord := GetLatestPowerRecord(buy.UserID)
	stocked := brecord.KwhStocked + txn.Kwh
	buyPR := PowerRecord{0, 0, stocked,
		brecord.KwhSaleable, now, brecord.MeterID, brecord.UserID}

	if DB.Create(&buyPR).Error != nil {
		return errors.New("error while updating Power Record of Meter ID[" + buyPR.MeterID + "]")
	}

	srecord := GetLatestPowerRecord(sell.UserID)
	saleable := srecord.KwhSaleable - txn.Kwh
	sellPR := PowerRecord{0, 0, srecord.KwhStocked,
		saleable, now, srecord.MeterID, srecord.UserID}
	if DB.Create(&sellPR).Error != nil {
		return errors.New("error while updating Power Record of Meter ID[" + sellPR.MeterID + "]")
	}

	return nil
}

func UpdateOrderByTxn(txn DealTxn) error {
	var buy, sell Order
	DB.First(&buy, "id = ?", txn.BuyOrderID)
	DB.First(&sell, "id = ?", txn.SellOrderID)

	fmt.Printf("BuyOrder: %+v\n", buy)
	fmt.Printf("SellOrder: %+v\n", sell)

	buy.KwhDealt = buy.KwhDealt + txn.Kwh
	if buy.Kwh == buy.KwhDealt {
		buy.Status = 2
	}
	if DB.Save(&buy).Error != nil {
		return errors.New("error while updating Buy Order[" + buy.ID + "]")
	}

	sell.KwhDealt = sell.KwhDealt + txn.Kwh
	if sell.Kwh == sell.KwhDealt {
		sell.Status = 2
	}
	if DB.Save(&sell).Error != nil {
		return errors.New("error while updating Sell Order[" + sell.ID + "]")
	}
	return nil
}

func AllUsers(page int, pagination int) (users []User, count int) {
	//limit = rows for each page
	//page 1 -> offset = 0*pagination
	//page 2 -> offset = 1*pagination
	DB.Model(&User{}).Count(&count)
	DB.Offset((page - 1) * pagination).Limit(pagination).Find(&users)
	return users, count
}

func LoginCheck(account string, passwd string) (*User, bool) {
	user := new(User)
	DB.Where(&User{Account: account, Password: passwd}).First(user)
	if user.Account == "" {
		return user, false
	}
	fmt.Println(user.Account + " exists in database !!")
	return user, true
}

func GetUserPowerRecords(uid uint, page int, pagination int) (records []PowerRecord, count int) {
	var user User
	DB.First(&user, uid)

	DB.Where("user_id = ?", uid).Model(&PowerRecord{}).Count(&count)
	//var records []PowerRecord
	DB.Offset((page - 1) * pagination).Limit(pagination).Model(&user).Related(&records)
	return records, count
}

func GetUserUndealtOrders(uid uint, tpe uint8, page int, pagination int) (orders []Order, count int) {
	var user User
	DB.First(&user, uid)

	DB.Where("user_id = ? AND status = ? AND type = ?", uid, 1, tpe).Model(&Order{}).Count(&count)
	DB.Offset((page-1)*pagination).Limit(pagination).Where("status = ? AND type = ?", 1, tpe).Model(&user).Related(&orders)
	return orders, count
}

func GetUserMeters(uid uint) (meters []MeterDeposit) {
	DB.Where("user_id = ?", uid).Find(&meters)
	return
}

func GetUserAllUndealtOrders(uid uint, tpe uint8) (orders []Order) {
	var user User
	DB.First(&user, uid)

	DB.Where("status = ? AND type = ?", 1, tpe).Model(&user).Related(&orders)
	return orders
}

func GetLatestPowerRecord(uid uint) (prd PowerRecord) {
	//var prd PowerRecord

	//now := time.Now().UTC().Add(8*time.Hour)
	local, _ := time.LoadLocation("Local")
	now := time.Now().In(local)
	DB.Order("updated_at desc").Where("user_id = ? AND updated_at < ? ", uid, now).First(&prd)
	//fmt.Printf("%+v\n", prd)

	return
}

func GetLatestMeterDeposit(orgId uint) interface{} {
	var meter MeterDeposit
	//now := time.Now().UTC().Add(8*time.Hour)
	local, _ := time.LoadLocation("Local")
	now := time.Now().In(local)
	DB.Order("updated_at desc").Where("org_id = ? AND updated_at < ? ", orgId, now).First(&meter)
	//fmt.Printf("%+v\n", prd)
	if meter.MeterID == "" {
		return nil
	}
	return meter
}

func GetSelectedMeterDeposit(uid uint) (meter MeterDeposit) {
	DB.Where("user_id = ? AND selected = true", uid).Find(&meter)

	return
}
