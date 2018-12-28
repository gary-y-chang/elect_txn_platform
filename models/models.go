package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"fmt"
	"time"
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

type Order struct {
	ID         string     // UserID-Type-timestamp-of-CreatedAt  timestamp := time.Now().Unix()
	Type       uint8      //1:buy, 2:sell
	Kwh        float64
    Price      float64
    CreatedAt  time.Time
    ExpiredAt  time.Time
    Status     uint8
	UserID     uint
}

var DB *gorm.DB

func init()  {
	var err error
	DB, err = gorm.Open("postgres", "host=192.168.1.4 port=5432 user=postgres dbname=platform_db password=postgres sslmode=disable")
	//DB, err = gorm.Open("postgres", "host=pgdb port=5432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	DB.LogMode(true)
	DB.SingularTable(true)
	DB.AutoMigrate(&User{}, &UserAttribute{}, &PowerRecord{})
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
