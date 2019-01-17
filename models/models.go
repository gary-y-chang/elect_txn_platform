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
	Account       string           `gorm:"unique;not null" json:"account"`
	Password      string           `gorm:"not null" json:"password"`
	Name          string           `json:"name"`
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
  "userid": 2
}
 **/
type Order struct {
	ID         string      `json:"id"`   // UUID  timestamp := time.Now().Unix()
	Type       uint8       `json:"type"`     //1:buy, 2:sell
	Kwh        float64     `json:"kwh"`
    Price      float64     `json:"price"`
    CreatedAt  time.Time   `json:"created"`
    ExpiredAt  time.Time   `json:"expired"`
    Status     uint8       `json:"status"`   //1:un-dealt, 2:dealt, 3:expired, 4:canceled
	UserID     uint        `json:"userid"`
}

var DB *gorm.DB

func init()  {
	var err error
	DB, err = gorm.Open("postgres", "host=192.168.1.4 port=15432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	//DB, err = gorm.Open("postgres", "host=pgdb port=5432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	DB.LogMode(true)
	DB.SingularTable(true)
	DB.AutoMigrate(&User{}, &UserAttribute{}, &PowerRecord{}, &Order{})
}

func AddUser(u User) error {

	if DB.Create(&u).Error != nil {
		return errors.New("user already exists")
	}
	return nil
}

func AllUsers() (users []User){
	//var users []User
	DB.Find(&users)
	return users
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

func AddPowerRecord(pr PowerRecord) error {
	//t := time.Now()
	//pr.ID = t.Format("2006-01-02")
	//pr.UpdatedAt = time.Now()
	if DB.Create(&pr).Error != nil {
		return errors.New("error while adding PowerRecord")
	}
	return nil
}

func GetUserPowerRecords(uid uint) (records []PowerRecord) {
	var user User
    DB.First(&user, uid)

	//var records []PowerRecord
	DB.Model(&user).Related(&records)
	return  records
}

func GetUserUndealtOrders(uid uint, tpe uint8) (orders []Order) {
	var user User
	DB.First(&user, uid)

	//var records []PowerRecord
	DB.Where("status = ? AND type = ?", 1, tpe).Model(&user).Related(&orders)
	return  orders
}

func AddOrder(odr Order) error {
	if DB.Create(&odr).Error != nil {
		return errors.New("error while adding Order")
	}
	return nil
}
