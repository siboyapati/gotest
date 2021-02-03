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

func Globalxetfs(mSync *sync.WaitGroup) {
	defer mSync.Done()
	m := make(map[string]string)
	var wg sync.WaitGroup
	m["CLOU"] = "https://www.globalxetfs.com/funds/clou/?download_full_holdings=true"
	m["GNOM"] = "https://www.globalxetfs.com/funds/gnom/?download_full_holdings=true"
	for fund, url := range m {
		wg.Add(1)
		go parseCSV(fund, url, &wg)
	}
	wg.Wait()
}

func parseCSV(fund, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	var timeStamp time.Time
	recordFile, err := http.Get(url) //os.Open("../Downloads/ark.csv")
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
		if i == 0 || i == 2 {
			continue
		}
		if i == 1 {
			strArr := strings.Split(line[0], " ")
			dateTime := strArr[len(strArr)-1]
			dateTime = strings.ReplaceAll(dateTime, "/", "-")
			timeStamp, err = time.Parse("1-2-2006", dateTime)
			if err != nil {
				log.Fatal(err)
			}
			//if the record already exists
			if DB.HasRecord(timeStamp, fund) {
				fmt.Println("Record alrady exists ###", fund)
				return
			}
			continue
		}
		weightString, ticker, company, cusip := line[0], line[1], line[2], line[3]
		if len(ticker) == 0 || len(cusip) == 0 || len(company) == 0 || len(cusip) == 0 {
			continue
		}
		closePriceString := line[4]
		sharesString := line[5]
		marketValueString := line[6]

		if len(ticker) == 0 || len(cusip) == 0 || len(closePriceString) == 0 || len(sharesString) == 0 {
			continue
		}
		weight, err := strconv.ParseFloat(weightString, 64)
		closePrice, err := strconv.ParseFloat(closePriceString, 64)
		sharesString = strings.ReplaceAll(sharesString, ",", "")
		shares, err := strconv.ParseFloat(sharesString, 64)

		marketValueString = strings.ReplaceAll(marketValueString, ",", "")
		marketValue, err := strconv.ParseFloat(marketValueString, 64)
		if err != nil {
			log.Fatal()
		}
		obj := model.Stock{
			Date:        timeStamp,
			Fund:        fund,
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

	if len(dict) == 0{
		return
	}
	DB.DailyETFHoldingDBSave(dict, timeStamp, fund)
	//result := DB.DailyChanges(fund)
	//fmt.Println("Updated daily changes for %s =", fund, result)
}
