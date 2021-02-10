package main

import (
	"encoding/csv"
	"fmt"
	"github.com/siboyapati/arketf/ETF"

	//"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/siboyapati/arketf/DB"
	"github.com/siboyapati/arketf/model"
	"sync"
	//"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func task() {
	fmt.Println("I am running task.")
}

func main() {

	//fmt.Println(utils.GetCurrentTime())
	//los, _ := time.LoadLocation("America/Los_Angeles")
	//c := cron.New(cron.WithLocation(los))
	//*/15 * * * *
	//0 16 * * ?
	//0 20 * * 1,3
	//*/5 9-17 * * 1-5
	//c.AddFunc("CRON_TZ=America/Los_Angeles */40 21-22 * * 1,5", func() {
	//	fmt.Println("Run every day at 6:00 PM ####", utils.GetCurrentTime())
	//})
	//c.AddFunc("@every 1m", task)
	//c.AddFunc("CRON_TZ=America/Los_Angeles 18 14 * * *", func() { fmt.Println("Runs at 04:30 Tokyo time every day") })
	//c.AddFunc("@every 1h30m", func() { fmt.Println("Every hour thirty, starting an hour thirty from now") })
	//c.AddFunc("1 * * * *", func() { fmt.Println("Every sec!!!!") })

	//c.AddFunc("*/5 * * * * * ", func() {
	//	log.Println("hello word", utils.GetCurrentTime())
	//})

	//
	//
	///
	//h := Hello{"I Love You!"}
	//Add a timed task
	//c.AddFunc("*/2 * * * * * ", func() {
	//	fmt.Println("hello1")
	//})
	//// Add a timed task
	//c.AddFunc("*/5 * * * * * ", func() {
	//	log.Println("hello word")
	//})
	//
	//c.AddFunc("1 * * * *", func() { fmt.Println("Every hour on the half hour####") })
	//c.Start()
	//
	//s, err := cron.Parse("*/3 * * * * *")
	//if err != nil {
	//	log.Println("Parse error")
	//}
	//h2 := Hello{"I Hate You!"}
	//c.Schedule(s, h2)
	//
	///

	//defer c.Stop()

	log.SetFlags(log.Lshortfile)
	start := time.Now()
	parseETF()
	fmt.Println("Time taken to parse all the etfs:", time.Since(start))
	select {}
}

func parseETF() {
	var mSync sync.WaitGroup

	mSync.Add(1)
	go parseArkErf(&mSync)

	mSync.Add(1)
	go ETF.IpoETF(&mSync)

	mSync.Add(1)
	go ETF.Globalxetfs(&mSync)

	mSync.Add(1)
	go ETF.WisdomCloudETF(&mSync)

	mSync.Add(1)
	go ETF.SPCX(&mSync)

	mSync.Add(1)
	go ETF.DefianceETFs(&mSync)

	mSync.Add(1)
	go ETF.SpxzETF(&mSync)

	mSync.Wait()

	//allETFS := getAllETFS()
	//var w sync.WaitGroup
	//for key, _ := range allETFS {
	//	w.Add(1)
	//	result := DB.DailyChanges(key)
	//	fmt.Println("Updated daily changes for %s =", key, result)
	//}
	//w.Wait()
}

func getAllETFS() map[string]bool {
	m := make(map[string]bool)
	m["ARKF"] = true
	m["ARKG"] = true
	m["ARKK"] = true
	m["ARKQ"] = true
	m["ARKW"] = true
	m["IPOETF"] = true
	m["CLOU"] = true
	m["GNOM"] = true
	m["WCLD"] = true
	return m
}

func parseArkErf(mSync *sync.WaitGroup) {
	defer mSync.Done()
	var wg sync.WaitGroup
	m := make(map[string]string)
	m["ARKF"] = "https://ark-funds.com/wp-content/fundsiteliterature/csv/ARK_FINTECH_INNOVATION_ETF_ARKF_HOLDINGS.csv"
	m["ARKG"] = "https://ark-funds.com/wp-content/fundsiteliterature/csv/ARK_GENOMIC_REVOLUTION_MULTISECTOR_ETF_ARKG_HOLDINGS.csv"
	m["ARKK"] = "https://ark-funds.com/wp-content/fundsiteliterature/csv/ARK_INNOVATION_ETF_ARKK_HOLDINGS.csv"
	m["ARKQ"] = "https://ark-funds.com/wp-content/fundsiteliterature/csv/ARK_AUTONOMOUS_TECHNOLOGY_&_ROBOTICS_ETF_ARKQ_HOLDINGS.csv"
	m["ARKW"] = "https://ark-funds.com/wp-content/fundsiteliterature/csv/ARK_NEXT_GENERATION_INTERNET_ETF_ARKW_HOLDINGS.csv"
	fmt.Println("starting Ark CSV Parser!!!")
	for key, value := range m {
		wg.Add(1)
		go csvReader(key, value, &wg)
	}
	wg.Wait()
}

func csvReader(etf, url string, wg *sync.WaitGroup) {
	defer wg.Done()
	dict := map[string]model.Stock{}
	shouldVerifyTimeStamp := true
	var timeStamp time.Time
	reader := FileReader(url)
	var header []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(record) < 8 {
			continue
		}
		if header == nil {
			header = record
		} else {
			dateField, fund, company, ticker, cusip := strings.TrimSpace(record[0]), strings.TrimSpace(record[1]), strings.TrimSpace(record[2]), strings.TrimSpace(record[3]), strings.TrimSpace(record[4])
			shareRecord, marketRecord, weightRecord := strings.TrimSpace(record[5]), strings.TrimSpace(record[6]), strings.TrimSpace(record[7])
			if len(shareRecord) == 0 || len(marketRecord) == 0 || len(weightRecord) == 0 {
				continue
			}
			dateField = strings.ReplaceAll(dateField, "/", "-")
			timeStamp, err = time.Parse("1-2-2006", dateField)
			if err != nil {
				fmt.Println(err)
				log.Fatal(err)
			}
			// Check if the record exists; if it does return; if not insert into mongodb;
			if shouldVerifyTimeStamp {
				if DB.HasRecord(timeStamp, etf) {
					fmt.Println("Record alrady exists", etf)
					return
				}
				shouldVerifyTimeStamp = false
			}
			shares, err := strconv.ParseFloat(shareRecord, 64)
			marketValue, err := strconv.ParseFloat(marketRecord, 64)
			weight, err := strconv.ParseFloat(weightRecord, 64)
			if err != nil {
				fmt.Println(err)
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
			var key string
			if len(ticker) != 0 {
				key = ticker
			} else {
				continue
			}
			dict[key] = obj
		}
	}
	if len(dict) == 0 {
		return
	}
	DB.DailyETFHoldingDBSave(dict, timeStamp, etf)
	fmt.Println("ARK parser success!!!", etf)
	//result := DB.DailyChanges(etf)
	//fmt.Println("Updated daily changes diff calculations for %s =", etf, result)
}

func FileReader(url string) *csv.Reader {
	recordFile, err := http.Get(url)
	if err != nil {
		fmt.Println("An error encountered ::", err)
		log.Fatal(err)
	}
	reader := csv.NewReader(recordFile.Body)
	reader.FieldsPerRecord = -1
	return reader
}
