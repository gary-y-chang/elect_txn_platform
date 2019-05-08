package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Account    string `gorm:"unique;not null"`
	Password   string `gorm:"not null"`
	Name       string
	Attributes []UserAttribute
}

type UserAttribute struct {
	UserID  uint
	Email   string
	Address string
	Phone   string
}

func (u *User) AddUser() error {
	if DB.Create(u).Error != nil {
		return errors.New("user already exists")
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
