package ETF

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/siboyapati/arketf/DB"
	"github.com/siboyapati/arketf/model"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func WisdomCloudETF(mSync *sync.WaitGroup) {
	defer mSync.Done()
	// Request the HTML page.
	res, err := http.Get("https://www.wisdomtree.com/global/etf-details/modals/all-current-day-holdings?id={EE8E5F82-67E0-4B72-8582-517D48D0AEF9}")
	dict := map[string]model.Stock{}
	var timeStamp time.Time

	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".timestamp span").Each(func(i int, s *goquery.Selection) {
		if i == 1 {
			docTimeStamp := s.Text()
			docTimeStamp = strings.ReplaceAll(docTimeStamp, "/", "-")
			timeStamp, err = time.Parse("1-2-2006", docTimeStamp)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	if DB.HasRecord(timeStamp, "WCLD") {
		fmt.Println("Record alrady exists WCLD")
		return
	}

	doc.Find(".table tbody").Each(func(i int, s *goquery.Selection) {
		var company string
		var ticker string
		var cusip string
		var shares float64
		var weight float64
		s.Find("tr").Each(func(index int, s1 *goquery.Selection) {
			s1.Find("td").Each(func(index int, s2 *goquery.Selection) {
				fieldValue := s2.Text()
				if index == 0 {
					companyArr := strings.Split(fieldValue, ".")
					company = companyArr[len(companyArr)-1]
				} else if index == 1 {
					tickerArr := strings.Split(fieldValue, " ")
					ticker = tickerArr[0]
				}
				if index == 2 {
					cusip = fieldValue
				}
				if index == 4 {
					shares, err = strconv.ParseFloat(fieldValue, 64)
					if err != nil {
						log.Fatal(err)
					}
				}
				if index == 5 {
					weightArr := strings.Split(fieldValue, "%")
					weight, err = strconv.ParseFloat(weightArr[0], 64)
					if err != nil {
						log.Fatal(err)
					}
				}
			})
			obj := model.Stock{
				Date:    timeStamp,
				Fund:    "WCLD",
				Company: company,
				Ticker:  ticker,
				Cusip:   cusip,
				Shares:  shares,
				Weight:  weight,
			}
			if len(ticker) != 0 {
				dict[ticker] = obj
			}
		})

	})
	if len(dict) == 0 {
		return
	}
	DB.DailyETFHoldingDBSave(dict, timeStamp, "WCLD")
	//result := DB.DailyChanges("WCLD")
	//fmt.Println("Updated daily changes for WCLD", result)
}
