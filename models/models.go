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
	ID           string  //yyyy-mm-dd
	KwhProduced  float64
	KwhConsumed  float64
	KwhSell      float64
	KwhBuy       float64
	Update       time.Time
	UserID       uint
}

type Order struct {

}

var DB *gorm.DB

func init()  {
	var err error
	DB, err = gorm.Open("postgres", "host=192.168.1.4 port=5432 user=postgres dbname=platform_db password=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	DB.LogMode(true)
	DB.SingularTable(true)
	DB.AutoMigrate(&User{}, &UserAttribute{}, &PowerRecord{})
}

func AddUser(u User) error {

	if DB.Create(&u).Error != nil {
		return errors.New("User already exists. ")
	}
	return nil
}

func AllUsers() []User{
	var users []User
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
	t := time.Now()
	pr.ID = t.Format("2006-01-02")
	pr.Update = t
	if DB.Create(&pr).Error != nil {
		return errors.New("Error while adding PowerRecord.")
	}
	return nil
}
