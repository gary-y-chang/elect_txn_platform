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
	MeterID     string `gorm:"not null"`
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
	Money     float64
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
	//DB, err = gorm.Open("postgres", "host=128.199.88.117 port=15432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	//DB, err = gorm.Open("postgres", "host=192.168.1.4 port=15432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	DB, err = gorm.Open("postgres", "host=pgdb port=5432 user=platformer dbname=platform_db password=postgres sslmode=disable")
	if err != nil {
		panic(err)
	}

	DB.LogMode(true)
	DB.SingularTable(true)
	DB.AutoMigrate(&User{}, &UserAttribute{}, &PowerRecord{}, &Order{}, &OrderBoard{}, &DealTxn{}, &DepositRecord{}, &MeterDeposit{})
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

func AddOrderBoard(ob OrderBoard) error {
	if DB.Create(&ob).Error != nil {
		return errors.New("error while adding OrderBoard")
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

func AddPowerRecord(pr *PowerRecord) (*PowerRecord, error) {
	lastPR := GetLatestPowerRecord(pr.MeterID)
	pr.KwhStocked = lastPR.KwhStocked - pr.KwhConsumed
	pr.KwhSaleable = lastPR.KwhSaleable + pr.KwhProduced
	if DB.Create(pr).Error != nil {
		return pr, errors.New("error while adding PowerRecord")
	}
	return pr, nil
}

func UpdatePowerRecordByTxn(txn DealTxn) error {
	// local, _ := time.LoadLocation("Local")
	// now := time.Now().In(local)
	now := time.Now().UTC().Add(8 * time.Hour)
	var buy, sell Order
	DB.First(&buy, "id = ?", txn.BuyOrderID)
	DB.First(&sell, "id = ?", txn.SellOrderID)
	mtBuy := GetMeterDepositByDepositNo(buy.DepositNo)
	brecord := GetLatestPowerRecord(mtBuy.MeterID)
	stocked := brecord.KwhStocked + txn.Kwh
	buyPR := PowerRecord{0, 0, stocked,
		brecord.KwhSaleable, now, brecord.MeterID}

	if DB.Create(&buyPR).Error != nil {
		return errors.New("error while updating Power Record of Meter ID[" + buyPR.MeterID + "]")
	}

	mtSell := GetMeterDepositByDepositNo(sell.DepositNo)
	srecord := GetLatestPowerRecord(mtSell.MeterID)
	saleable := srecord.KwhSaleable - txn.Kwh
	sellPR := PowerRecord{0, 0, srecord.KwhStocked,
		saleable, now, srecord.MeterID}
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

func LoginCheck(account string, passwd string) (User, bool) {
	user := User{}
	//DB.Where(&User{Account: account, Password: passwd}).First(&user)
	DB.Where("account = ? AND password = ?", account, passwd).Find(&user)
	if user.Account == "" {
		return user, false
	}
	fmt.Println(user.Account + " exists in database !!")
	return user, true
}

func GetPowerRecordsByMeter(mid string, page int, pagination int) (records []PowerRecord, count int) {
	//var user User
	//DB.First(&user, uid)

	DB.Where("meter_id = ?", mid).Model(&PowerRecord{}).Count(&count)
	//var records []PowerRecord
	DB.Offset((page-1)*pagination).Limit(pagination).Where("meter_id = ?", mid).Find(&records)
	return records, count
}

func GetUndealtOrdersByMeter(mid string, tpe uint8, page int, pagination int) (orders []Order, count int) {
	DB.Where("meter_id = ? AND status = ? AND type = ?", mid, 1, tpe).Model(&Order{}).Count(&count)
	DB.Offset((page-1)*pagination).Limit(pagination).Where("meter_id = ? AND status = ? AND type = ?", mid, 1, tpe).Find(&orders)

	return orders, count
}

func GetUndealtKwhByMeter(mid string, tpe uint8) float64 {
	type Result struct {
		Undealt float64
	}
	var result Result
	DB.Table("order").Select("sum(kwh - kwh_dealt) as Undealt").Where("meter_id = ? AND status = ? AND type = ?", mid, 1, tpe).Scan(&result)
	//fmt.Printf("kwh: %f\n", result.Undealt)
	return result.Undealt
}

func GetUserMeters(uid uint) (meters []MeterDeposit) {
	DB.Where("user_id = ?", uid).Find(&meters)
	return
}

func GetUserAllUndealtOrders(uid uint, tpe uint8) (orders []Order) {
	// var user User
	// DB.First(&user, uid)

	// DB.Where("status = ? AND type = ?", 1, tpe).Model(&user).Related(&orders)
	return orders
}

func GetLatestPowerRecord(mid string) (prd PowerRecord) {
	//var prd PowerRecord

	now := time.Now().UTC().Add(8 * time.Hour)
	// local, _ := time.LoadLocation("Asia/Taipei")
	// now := time.Now().In(local)
	DB.Order("updated_at desc").Where("meter_id = ? AND updated_at < ?", mid, now).First(&prd)
	//fmt.Printf("%+v\n", prd)

	return
}

func GetLatestMeterDeposit(orgId uint) interface{} {
	var meter MeterDeposit
	now := time.Now().UTC().Add(8 * time.Hour)
	// local, _ := time.LoadLocation("Asia/Taipei")
	// now := time.Now().In(local)
	DB.Order("updated_at desc").Where("org_id = ? AND updated_at < ? ", orgId, now).First(&meter)
	//fmt.Printf("%+v\n", prd)
	if meter.MeterID == "" {
		return nil
	}
	return meter
}

func GetDefaultMeterDeposit(uid uint) (meter MeterDeposit) {
	DB.Where("user_id = ? AND is_default = true", uid).Find(&meter)
	return
}

func GetMeterDepositByID(mid string) (meter MeterDeposit) {
	DB.Where("meter_id = ?", mid).Find(&meter)
	return
}

func GetMeterDepositByDepositNo(depo_no string) (meter MeterDeposit) {
	DB.Where("deposit_no = ?", depo_no).Find(&meter)
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

func UpdateOrderLocked(oid string, txnKwh float64) error {
	var order Order
	tx := DB.Begin()
	tx.Begin()

	notFound := DB.Raw("SELECT * FROM public.order_board WHERE id=? FOR UPDATE;", oid).Scan(&order).RecordNotFound()
	if notFound {
		return errors.New("order: " + oid + " already closed")
	}

	order.KwhDealt = order.KwhDealt + txnKwh
	if order.Kwh == order.KwhDealt {
		tx.Exec("DELETE FROM public.order_board WHERE id = ?", oid)
	} else {
		tx.Exec("UPDATE public.order_board SET kwh_dealt=? WHERE id = ?", order.KwhDealt, oid)
	}

	tx.Commit()
	//fmt.Printf("%+v\n", order)
	//DB.First(&order, "id = ?", oid)
	return nil
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
