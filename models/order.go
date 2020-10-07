package models

import (
	"time"
	"errors"
)

type Order struct {
	ID        string  // UUID           timestamp := time.Now().Unix()
	Type      uint8   // 1:buy, 2:sell
	Kwh       float64 // Kwh to be sold
	KwhDealt  float64 // Kwh already sold
	Price     float64
	CreatedAt time.Time
	ExpiredAt time.Time
	Status    uint8 //1:un-dealt, 2:dealt, 3:expired, 4:canceled
	MeterID   string
	DepositNo string
}

type OrderBoard struct {
	ID        string  // UUID           timestamp := time.Now().Unix()
	Type      uint8   // 1:buy, 2:sell
	Kwh       float64 // Kwh to be sold
	KwhDealt  float64 // Kwh already sold
	Price     float64
	CreatedAt time.Time
	ExpiredAt time.Time
	Status    uint8 //1:un-dealt, 2:dealt, 3:expired, 4:canceled
	MeterID   string
	DepositNo string
}

type OrderDisplay struct {
	ID        string  // UUID           timestamp := time.Now().Unix()
	Type      uint8   // 1:buy, 2:sell
	Kwh       float64 // Kwh to be sold
	KwhDealt  float64 // Kwh already sold
	Price     float64
	CreatedAt string
	ExpiredAt string
	Status    uint8 //1:un-dealt, 2:dealt, 3:expired, 4:canceled
	MeterID   string
	DepositNo string
}

func (od *OrderDisplay) Transmit(order Order) OrderDisplay {
	od.ID = order.ID
	od.Type = order.Type
	od.Price = order.Price
	od.Kwh = order.Kwh
	od.KwhDealt = order.KwhDealt
	od.Status = order.Status
	od.MeterID = order.MeterID
	od.DepositNo = order.DepositNo
	od.CreatedAt = order.CreatedAt.Format("2006-01-02 15:04:05")
	od.ExpiredAt = order.ExpiredAt.Format("2006-01-02 15:04:05")

	return *od
}

func AddOrder(odr Order) error {
	if DB.Create(&odr).Error != nil {
		return errors.New("error while adding Order")
	}
	return nil
}

func AddOrderBoard(ob OrderBoard) error {
	if DB.Create(&ob).Error != nil {
		return errors.New("error while adding OrderBoard")
	}
	return nil
}

func GetUndealtOrdersByMeter(mid string, tpe uint8, page int, pagination int) (odrs []OrderDisplay, count int) {
	var orders []Order
	DB.Where("meter_id = ? AND status = ? AND type = ?", mid, 1, tpe).Model(&Order{}).Count(&count)
	DB.Offset((page-1)*pagination).Limit(pagination).Where("meter_id = ? AND status = ? AND type = ?", mid, 1, tpe).Find(&orders)
	
	for _, o := range orders {
		od := new(OrderDisplay)
		odrs = append(odrs, od.Transmit(o))
	}
	return
}

func QueryOrders(tpe uint8, status uint8, mid string, begin time.Time, end time.Time, page int, pagination int) (orders []Order, count int) {
	if status == 0 {
		DB.Where("type = ? AND meter_id = ? AND date(created_at) BETWEEN ? AND ?", tpe, mid, begin, end).Model(&Order{}).Count(&count)
		DB.Offset((page-1)*pagination).Limit(pagination).Where("type = ? AND meter_id = ? AND date(created_at) BETWEEN ? AND ?", tpe, mid, begin, end).Find(&orders)
	} else {
		DB.Where("type = ? AND status = ? AND meter_id = ? AND date(created_at) BETWEEN ? AND ?", tpe, status, mid, begin, end).Model(&Order{}).Count(&count)
		DB.Offset((page-1)*pagination).Limit(pagination).Where("type = ? AND status = ? AND meter_id = ? AND date(created_at) BETWEEN ? AND ?", tpe, status, mid, begin, end).Find(&orders)
	}

	return
}

func GetUndealtOrders(tpe uint8) (orders []OrderBoard) {
	if tpe == 1 {
		DB.Order("price desc").Order("created_at asc").Where("status = ? AND type = ? AND expired_at > NOW()", 1, tpe).Find(&orders)
	} else if tpe == 2 {
		DB.Order("price asc").Order("created_at asc").Where("status = ? AND type = ? AND expired_at > NOW()", 1, tpe).Find(&orders)
	}

	return
}

func (ob *OrderBoard) Transmit(order Order) OrderBoard {
	ob.ID = order.ID
	ob.Type = order.Type
	ob.Price = order.Price
	ob.Kwh = order.Kwh
	ob.KwhDealt = order.KwhDealt
	ob.Status = order.Status
	ob.MeterID = order.MeterID
	ob.DepositNo = order.DepositNo
	ob.CreatedAt = order.CreatedAt
	ob.ExpiredAt = order.ExpiredAt

	return *ob
}

