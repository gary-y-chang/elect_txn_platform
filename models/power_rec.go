package models

import (
	"errors"
	"time"
	"fmt"
	"sort"
	"math"
)

// type PowerRecord struct {
// 	KwhProduced float64 // power kwh produced and saved to 儲能櫃 during a time period, not accumulated
// 	KwhConsumed float64 // power kwh consumed from 儲能櫃 during a time period, not accumulated
// 	KwhStocked  float64 // power kwh accumulated from buy-order, stocked = stocked - consumed
// 	KwhSaleable float64 // power kwh accumulated from produced, saleable = saleable + produced
// 	UpdatedAt   time.Time
// 	MeterID     string `gorm:"not null"`
// }

type PowerStocked struct {
	KwhConsumed float64 // power kwh consumed from 儲能櫃 during a time period, not accumulated
	KwhStocked  float64 // power kwh accumulated from buy-order, stocked = stocked - consumed
	UpdatedAt   time.Time
	MeterID     string `gorm:"not null"`
}

type PowerSaleable struct {
	KwhProduced float64 // power kwh produced and saved to 儲能櫃 during a time period, not accumulated
	KwhSaleable float64 // power kwh accumulated from produced, saleable = saleable + produced
	UpdatedAt   time.Time
	MeterID     string `gorm:"not null"`
}

type PowerAnalysis struct {
	Date string
	Produced float64
	Consumed float64
	Sold float64
	Bought float64
}

type PowerSummary struct {
	TotalProduced float64
	TotalConsumed float64
	AvgProduced float64
	AvgConsumed float64
	TotalSold float64
	TotalBought float64
	Analyses []PowerAnalysis
}

func AddPowerRecord(produce float64, consume float64, mid string) error {
	now := time.Now().UTC().Add(8 * time.Hour)
	
	last := GetLatestPowerStocked(mid)
	stock := last.KwhStocked - consume 
	pwrt := PowerStocked{consume, stock, now, mid}
	if DB.Create(&pwrt).Error != nil {
		return errors.New("error while updating Power Stocked of Meter ID[" + mid + "]")
	}

	lasts := GetLatestPowerSaleable(mid)
	saleable := lasts.KwhSaleable + produce
	pwrs := PowerSaleable{produce, saleable, now, mid}
	if DB.Create(&pwrs).Error != nil {
		return errors.New("error while updating Power Saleable of Meter ID[" + mid + "]")
	}

	return nil
 }

func GetLatestPowerSaleable(mid string) (pwrs PowerSaleable) {
	now := time.Now().UTC().Add(8 * time.Hour)
	DB.Order("updated_at desc").Where("meter_id = ? AND updated_at < ?", mid, now).First(&pwrs)
	//fmt.Printf("%+v\n", pwrs)
	return
}

func GetLatestPowerStocked(mid string) (pwrt PowerStocked) {
	now := time.Now().UTC().Add(8 * time.Hour)
	DB.Order("updated_at desc").Where("meter_id = ? AND updated_at < ?", mid, now).First(&pwrt)
	//fmt.Printf("%+v\n", pwrt)
	return
}

func UpdatePowerRecordByTxn(txn DealTxn) error {

	now := time.Now().UTC().Add(8 * time.Hour)
	//var buy, sell Order
	// DB.First(&buy, "id = ?", txn.BuyOrderID)
	// DB.First(&sell, "id = ?", txn.SellOrderID)
	// mtBuy := GetMeterDepositByDepositNo(buy.DepositNo)

	last := GetLatestPowerStocked(txn.BuyMeterID)
	stock := last.KwhStocked + txn.Kwh
	pwrt := PowerStocked{0, stock, now, txn.BuyMeterID}
	// buyPR := PowerRecord{0, 0, stocked,
	// 	brecord.KwhSaleable, now, brecord.MeterID}

	if DB.Create(&pwrt).Error != nil {
		return errors.New("error while updating Power Stocked of Meter ID[" + txn.BuyMeterID + "]")
	}

	lasts := GetLatestPowerSaleable(txn.SellMeterID)
	saleable := lasts.KwhSaleable - txn.Kwh
	pwrs := PowerSaleable{0, saleable, now, txn.SellMeterID}
	if DB.Create(&pwrs).Error != nil {
		return errors.New("error while updating Power Saleable of Meter ID[" + txn.SellMeterID + "]")
	}

	return nil
}


func GetPowerAnalysis(mid string, begin time.Time, end time.Time, page int, pagination int) (PowerSummary, int){
	// return   date | kwh_produce | kwh_consume | txn_sell | txn_buy
	
	analysis := make(map[time.Time]PowerAnalysis) // map["2019-02-15"] = PowerAnalysis
	
	rows, _ := DB.Offset((page - 1) * pagination).Limit(pagination).Table("power_saleable").Select("sum(kwh_produced) as produced, date(updated_at) as date").Where("meter_id = ? AND date(updated_at) Between ? AND ?", mid, begin, end).Group("date(updated_at)").Rows()
	defer rows.Close()
	
	for rows.Next() {
		var produced float64
		var date time.Time
		rows.Scan(&produced, &date)
		pa, _ := analysis[date]
		pa.Produced = produced
		pa.Date = date.Format("2006-01-02")
		analysis[date] = pa
		fmt.Printf("Produced: %+v on %s \n", pa, date.Format("2006-01-02"))
	}

	rows, _ = DB.Offset((page - 1) * pagination).Table("power_stocked").Select("sum(kwh_consumed) as consumed, date(updated_at) as date").Where("meter_id = ? AND date(updated_at) Between ? AND ?", mid, begin, end).Group("date(updated_at)").Rows()
	for rows.Next() {
		var consumed float64
		var date time.Time
		rows.Scan(&consumed, &date)
		pa, _ := analysis[date]
		pa.Consumed = consumed
		pa.Date = date.Format("2006-01-02")
		analysis[date] = pa
		fmt.Printf("==> %s : %+v \n", date.Format("2006-01-02"), pa)
	}

	rows, _ = DB.Offset((page - 1) * pagination).Table("deal_txn").Select("sum(kwh), date(txn_date) as date").Where("sell_meter_id = ? AND date(txn_date) Between ? AND ?", mid, begin, end).Group("date(txn_date)").Rows()
	for rows.Next() {
		var sold float64
		var date time.Time
		rows.Scan(&sold, &date)
		pa, _ := analysis[date]
		pa.Sold = sold
		pa.Date = date.Format("2006-01-02")
		analysis[date] = pa
		fmt.Printf("==> %s : %+v \n", date.Format("2006-01-02"), pa)
	}

	rows, _ = DB.Offset((page - 1) * pagination).Table("deal_txn").Select("sum(kwh), date(txn_date) as date").Where("buy_meter_id = ? AND date(txn_date) Between ? AND ?", mid, begin, end).Group("date(txn_date)").Rows()
	for rows.Next() {
		var bought float64
		var date time.Time
		rows.Scan(&bought, &date)
		pa, _ := analysis[date]
		pa.Bought = bought
		pa.Date = date.Format("2006-01-02")
		analysis[date] = pa
		fmt.Printf("==> %s : %+v \n", date.Format("2006-01-02"), pa)
	}

	var totalProduced, totalConsumed, totalSold, totalBought float64
	days := []time.Time{}
	for k, v := range analysis {
		fmt.Println(v)
		days = append(days, k)
		totalProduced += v.Produced
		totalConsumed += v.Consumed
		totalSold += v.Sold
		totalBought += v.Bought
	} 
	fmt.Printf("==> %.2f  %.2f  %.2f  %.2f \n", totalProduced, totalConsumed, totalSold, totalBought)
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })
	//fmt.Println("Sorting By Days:", days)
	data := []PowerAnalysis{}
	for _, a := range days {
		data = append(data, analysis[a])
	}

	var duration, avgProduced, avgConsumed float64
	if len(days) > 0 {
		duration = (days[len(days)-1].Sub(days[0]).Hours() / 24) + 1
		avgProduced = math.Round((totalProduced / duration) * 100) / 100
		avgConsumed = math.Round((totalConsumed / duration) * 100) / 100
	}	
	// fmt.Printf("==> days = %.0f \n", duration)
	// avgProduced := math.Round((totalProduced / duration) * 100) / 100
	// avgConsumed := math.Round((totalConsumed / duration) * 100) / 100

	summary := PowerSummary{totalProduced, totalConsumed, avgProduced, avgConsumed, totalSold, totalBought, data}
	
	return summary, len(data)
}

