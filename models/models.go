package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
	"fmt"
	"errors"
)

type User struct {
	gorm.Model
	Account       string           `gorm:"unique;not null"`
	Password      string           `gorm:"not null"`
	Name          string
	Attributes    []UserAttribute
	PowerRecords  []PowerRecord
}

type UserAttribute struct {
	UserID   uint
	Email    string
	Address  string
	Phone    string
}

type PowerRecord struct {
	KwhProduced  float64
	KwhConsumed  float64
	KwhStocked   float64
	UpdatedAt    time.Time
	UserID       uint
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
	ID         string        // UUID           timestamp := time.Now().Unix()
	Type       uint8         // 1:buy, 2:sell
	Kwh        float64       // Kwh to be sold
	KwhDealt   float64       // Kwh already sold
    Price      float64
    CreatedAt  time.Time
    ExpiredAt  time.Time
    Status     uint8         //1:un-dealt, 2:dealt, 3:expired, 4:canceled
	UserID     uint
}

type DealTxn struct {
	ID           string     `gorm:"primary_key"` // transaction id from chaincode
	Part         uint8      `gorm:"primary_key"`
	Kwh          float64
	Price        float64    // dealt price
	TxnDate      time.Time  // display format 2018-11-20 23:45
	BuyOrderID   string     // order id
	SellOrderID  string     // order id
}

var DB *gorm.DB

func init()  {
	var err error
	//DB, err = gorm.Open("postgres", "host=192.168.43.214 port=15432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	//DB, err = gorm.Open("postgres", "host=192.168.1.4 port=15432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	DB, err = gorm.Open("postgres", "host=pgdb port=5432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	DB.LogMode(true)
	DB.SingularTable(true)
	DB.AutoMigrate(&User{}, &UserAttribute{}, &PowerRecord{}, &Order{}, &DealTxn{})
}

func AddUser(u User) error {

	if DB.Create(&u).Error != nil {
		return errors.New("user already exists")
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

func AddPowerRecord(pr PowerRecord) error {
	//t := time.Now()
	//pr.ID = t.Format("2006-01-02")
	//pr.UpdatedAt = time.Now()
	if DB.Create(&pr).Error != nil {
		return errors.New("error while adding PowerRecord")
	}
	return nil
}

func UpdateOrderByTxn(txn DealTxn ) error {
	var buy, sell Order
	DB.First(&buy, "id = ?", txn.BuyOrderID)
	DB.First(&sell, "id = ?", txn.SellOrderID)

	buy.KwhDealt = buy.KwhDealt + txn.Kwh
	if buy.Kwh == buy.KwhDealt {
		buy.Status = 2
	}
	if DB.Save(&buy).Error != nil {
		return errors.New("error while updating Buy Order["+buy.ID+"]")
	}

	sell.KwhDealt = sell.KwhDealt + txn.Kwh
	if sell.Kwh == sell.KwhDealt {
		sell.Status = 2
	}
	if DB.Save(&sell).Error != nil {
		return errors.New("error while updating Sell Order["+sell.ID+"]")
	}
	return nil
}

func AllUsers(page int, pagination int) (users []User, count int){
	//limit = rows for each page
	//page 1 -> offset = 0*pagination
	//page 2 -> offset = 1*pagination
	DB.Model(&User{}).Count(&count)
	DB.Offset((page-1)*pagination).Limit(pagination).Find(&users)
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
	DB.Offset((page-1)*pagination).Limit(pagination).Model(&user).Related(&records)
	return  records, count
}

func GetUserUndealtOrders(uid uint, tpe uint8, page int, pagination int) (orders []Order, count int) {
	var user User
	DB.First(&user, uid)

	DB.Where("user_id = ? AND status = ? AND type = ?", uid, 1, tpe).Model(&Order{}).Count(&count)
	DB.Offset((page-1)*pagination).Limit(pagination).Where("status = ? AND type = ?", 1, tpe).Model(&user).Related(&orders)
	return  orders, count
}

func GetUserAllUndealtOrders(uid uint, tpe uint8) (orders []Order) {
	var user User
	DB.First(&user, uid)

	DB.Where("status = ? AND type = ?", 1, tpe).Model(&user).Related(&orders)
	return  orders
}

func GetLatestPowerRecord(uid uint) (prd PowerRecord){
	//var prd PowerRecord

	now := time.Now().UTC().Add(8*time.Hour)
	DB.Order("updated_at desc").Where("user_id = ? AND updated_at < ? ", uid, now).First(&prd)
	fmt.Printf("%+v\n", prd)

	return
}


