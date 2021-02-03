package ETF

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/siboyapati/arketf/DB"
	"github.com/siboyapati/arketf/model"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func IpoETF(mSync *sync.WaitGroup) {
	defer mSync.Done()
	dict := map[string]model.Stock{}
	var time time.Time

	etf := "IPOETF"
	//shouldVerifyTimeStamp := true
	url := "https://www.renaissancecapital.com/IPO-Investing/US-IPO-ETF-Holdings-Excel-Download"
	recordFile, err := http.Get(url)
	if err != nil {
		fmt.Println("An error encountered ::", err)
		log.Fatal(err)
	}
	defer recordFile.Body.Close()

	f, err := excelize.OpenReader(recordFile.Body)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		return
	}
	rows, err := f.GetRows("Holdings")
	if err != nil {
		log.Fatal(err)
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		floatTime, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			log.Fatal(err)
		}
		time, err = excelize.ExcelDateToTime(floatTime, false)
		if err != nil {
			log.Fatal(err)
		}

		//if shouldVerifyTimeStamp {
		//	if DB.HasRecord(time, "IPOETF") {
		//		fmt.Println("Record alrady exists IPOETF")
		//		return
		//	}
		//	shouldVerifyTimeStamp = false
		//}
		company, ticker, cusip := row[1], row[3], row[4]

		if len(company) == 0 || len(cusip) == 0 || len(ticker) == 0 {
			continue
		}
		if len(row[5]) == 0 || len(row[6]) == 0 || len(row[7]) == 0 {
			continue
		}

		shares, _ := strconv.ParseFloat(row[5], 64)
		marketValue, _ := strconv.ParseFloat(row[6], 64)
		weight, _ := strconv.ParseFloat(row[7], 64)
		closePrice := marketValue / shares

		obj := model.Stock{
			Date:        time,
			Fund:        "IPO",
			Company:     company,
			Ticker:      ticker,
			Cusip:       cusip,
			Shares:      shares,
			MarketValue: marketValue,
			Weight:      weight,
			ClosePrice:  closePrice,
		}
		dict[ticker] = obj
	}
	if len(dict) == 0 {
		return
	}
	if DB.HasRecord(time, "IPOETF") {
		fmt.Println("Record alrady exists IPOETF")
		if DB.ShouldInsert(etf,dict){
			DB.UpdateDB(etf, dict)
		}
		return
	}
	DB.DailyETFHoldingDBSave(dict, time, "IPOETF")
	//result := DB.DailyChanges("IPOETF")
	//fmt.Println("Updated daily changes for IPOETF =", result)
}
