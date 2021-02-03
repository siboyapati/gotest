package ETF

import (
	"bytes"
	"fmt"
	"github.com/extrame/xls"
	"github.com/siboyapati/arketf/DB"
	"github.com/siboyapati/arketf/model"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func DefianceETFs(mSync *sync.WaitGroup) {
	defer mSync.Done()
	m := make(map[string]string)
	var wg sync.WaitGroup
	//m["IBBJ"] = "https://www.defianceetfs.com/download-holdings-usbanks.php?fund=IBBJ"
	m["SPAK"] = "https://www.defianceetfs.com/download-holdings-usbanks.php?fund=SPAK"
	for fund, url := range m {
		wg.Add(1)
		go excelParser(fund, url, &wg)
	}
	wg.Wait()
}

func excelParser(fund string, url string, mSync *sync.WaitGroup) {
	defer mSync.Done()

	loc, _ := time.LoadLocation("America/Los_Angeles")
	dayOfWeek := time.Now().In(loc).Weekday()

	//
	if dayOfWeek == 0 || dayOfWeek == 6 {
		fmt.Println("day of week", fund)
		return
	}

	dict := map[string]model.Stock{}
	var timeStamp time.Time
	//shouldVerifyTimeStamp := true
	recordFile, err := http.Get(url)
	if err != nil {
		fmt.Println("An error encountered ::", err)
		log.Fatal(err)
	}
	defer recordFile.Body.Close()
	body, err := ioutil.ReadAll(recordFile.Body)
	if err != nil {
		fmt.Println(err)
	}
	readSeeker := bytes.NewReader(body)
	xls, err := xls.OpenReader(readSeeker, "utf-8")
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		return
	}

	v := xls.ReadAllCells(10000)
	var company string
	var ticker string
	var cusip string
	var shares float64
	var marketValue float64
	var weight float64

	for id, col := range v {
		if col == nil || id == 0 || id == 2 || len(col) != 6{
			continue
		}
		if id == 1 {
			dateTime := strings.ReplaceAll(col[0], "/", "-")
			timeStamp, err = time.Parse("2006-01-02", dateTime)
			if DB.HasRecord(timeStamp, fund) {
				fmt.Println("Record alrady exists ###", fund)
				return
			}
			if err != nil {
				fmt.Println(err)
			}
			continue
		}

		//fmt.Println(id)
		for in, fieldValue := range col {
			if in == 0 {
				fieldValue = strings.Replace(fieldValue, "%", "", -1)
				weight, err = strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					weight = 0
					log.Println(err)
					//log.Fatal(err)
				}

			} else if in == 1 {
				company = fieldValue
			} else if in == 2 {
				ticker = fieldValue
			} else if in == 3 {
				cusip = fieldValue
			} else if in == 4 {
				fieldValue = strings.ReplaceAll(fieldValue, ",", "")
				shares, err = strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					log.Println(err)
					log.Fatal(err)
				}
			} else if in == 5 {
				fieldValue = strings.ReplaceAll(fieldValue, ",", "")
				marketValue, err = strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					log.Println(err)
					log.Fatal(err)
				}
			}
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
			ClosePrice:  marketValue / shares,
		}
		if len(ticker) != 0 {
			dict[ticker] = obj
		}
	}

	if len(dict) == 0 {
		return
	}
	DB.DailyETFHoldingDBSave(dict, timeStamp, fund)
}
