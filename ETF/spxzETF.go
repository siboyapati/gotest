package ETF

import (
	"encoding/csv"
	"fmt"
	"github.com/siboyapati/arketf/DB"
	"github.com/siboyapati/arketf/model"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SpxzETF(mSync *sync.WaitGroup) {
	defer mSync.Done()
	m := make(map[string]string)
	var wg sync.WaitGroup
	m["SPXZ"] = "https://morgancreeketf.filepoint.live/assets/data/FilepointMorganCreek.40VN.VN_Holdings.csv"
	for fund, url := range m {
		wg.Add(1)
		go parse(fund, url, &wg)
	}
	wg.Wait()
}

func parse(fund, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	shouldVerifyTimeStamp := true
	var timeStamp time.Time
	recordFile, err := http.Get(url)
	if err != nil {
		fmt.Println("An error encountered ::", err)
	}
	reader := csv.NewReader(recordFile.Body)
	reader.FieldsPerRecord = -1
	lines, err := reader.ReadAll()
	dict := map[string]model.Stock{}
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}


	for i, line := range lines {
		if i == 0 {
			continue
		}
		//if i == 1 {
		//	strArr := strings.Split(line[0], " ")
		//	dateTime := strArr[len(strArr)-1]
		//	dateTime = strings.ReplaceAll(dateTime, "/", "-")
		//	timeStamp, err = time.Parse("1-2-2006", dateTime)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	//if the record already exists
		//	if DB.HasRecord(timeStamp, fund) {
		//		fmt.Println("Record alrady exists ###", fund)
		//		return
		//	}
		//	continue
		//}

		//Date 0
		//etf 1
		//ticker 2
		//cusip 3
		//name 4
		//totalshares 5
		//price 6
		//marketValue 7
		//weight 8

		dateTime, ticker, cusip, company := line[0], line[2], line[3], line[4]
		if len(ticker) == 0 || len(cusip) == 0 || len(company) == 0  {
			continue
		}

		dateTime = strings.ReplaceAll(dateTime, "/", "-")
		timeStamp, err = time.Parse("1-2-2006", dateTime)
		if err != nil {
			log.Fatal(err)
		}

		shares := line[5]
		price := line[6]
		marketValue := line[7]
		weight := line[8]
		weight = strings.ReplaceAll(weight, "%", "")

		if len(shares) == 0 || len(price) == 0 || len(weight) == 0 || len(marketValue) == 0 {
			continue
		}

		totalShares, err := strconv.ParseFloat(shares, 64)
		if err != nil {
			log.Fatal(err)
		}
		closePrice, err := strconv.ParseFloat(price, 64)
		if err != nil {
			log.Fatal(err)
		}
		totalMarketValue, err := strconv.ParseFloat(marketValue, 64)
		if err != nil {
			log.Fatal(err)
		}
		weightPercentage, err := strconv.ParseFloat(weight, 64)

		if err != nil {
			log.Fatal(err)
		}

		if shouldVerifyTimeStamp {
			if DB.HasRecord(timeStamp, fund) {
				fmt.Println("Record alrady exists", fund)
				return
			}
			shouldVerifyTimeStamp = false
		}

		obj := model.Stock{
			Date:        timeStamp,
			Fund:        fund,
			Company:     company,
			Ticker:      ticker,
			Cusip:       cusip,
			Shares:      totalShares,
			MarketValue: totalMarketValue,
			Weight:      weightPercentage,
			ClosePrice:  closePrice,
		}
		dict[ticker] = obj
	}

	if len(dict) == 0 {
		return
	}

	DB.DailyETFHoldingDBSave(dict, timeStamp, fund)
	//result := DB.DailyChanges(fund)
	//fmt.Println("Updated daily changes for %s =", fund, result)
}
