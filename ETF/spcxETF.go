package ETF

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/siboyapati/arketf/DB"
	"github.com/siboyapati/arketf/model"
	"github.com/siboyapati/arketf/utils"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

func SPCX(mSync *sync.WaitGroup) {
	defer mSync.Done()
	dict := map[string]model.Stock{}
	etf := "SPCX"
	var timeStamp time.Time
	body := utils.ReadWebPage("https://www.spcxetf.com/spcx-holdings/")
	//defer body.Close()
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".et_pb_code_inner #tablepress-12 .row-hover .row-1 .column-2").Each(func(i int, s *goquery.Selection) {
		docTimeStamp := s.Text()
		docTimeStamp = strings.ReplaceAll(docTimeStamp, "/", "-")
		docTimeStampArr := strings.Split(docTimeStamp, "-")
		if len(docTimeStampArr) <= 2 {
			year := time.Now().UTC().Year()
			docTimeStampArr = append(docTimeStampArr, strconv.Itoa(year))
			docTimeStamp = strings.Join(docTimeStampArr, "-")
		}
		timeStamp, err = time.Parse("1-2-2006", docTimeStamp)

		//if the record already exists

		if err != nil {
			timeStamp = utils.GetCurrentTime()
			fmt.Println(err)
		}
	})

	if DB.HasRecord(timeStamp, etf) {
		fmt.Println("Record alrady exists ###", etf)
		return
	}

	//if DB.HasRecord(timeStamp, "SPCX") {
	//	fmt.Println("Record alrady exists WCLD")
	//	return
	//}
	//.et_pb_code_inner

	//type Stock struct {
	//	Date        time.Time
	//	Fund        string
	//	Company     string
	//	Ticker      string
	//	Cusip       string
	//	Shares      float64
	//	MarketValue float64
	//	Weight      float64
	//	ClosePrice  float64
	//}

	doc.Find("#tablepress-11").Each(func(i int, s *goquery.Selection) {
		var company string
		var ticker string
		var cusip string
		var shares float64
		var marketValue float64
		var weight float64

		s.Find("tbody tr").Each(func(trIndex int, s1 *goquery.Selection) {
			//fmt.Println(trIndex)
			s1.Find("td").Each(func(index int, s2 *goquery.Selection) {
				fieldValue := s2.Text()
				fieldValue = strings.TrimSpace(fieldValue)

				if index == 0 {
					if len(fieldValue) == 0 {
						return
					}
					ticker = fieldValue
					//companyArr := strings.Split(fieldValue, ".")
					//company = companyArr[len(companyArr)-1]
				} else if index == 1 {
					company = fieldValue
				} else if index == 2 {
					fieldValue = strings.Replace(fieldValue, "%", "", -1)
					weight, err = strconv.ParseFloat(fieldValue, 64)
					if err != nil {
						log.Println(err)
						log.Fatal(err)
					}
				} else if index == 3 {
					fieldValue = strings.ReplaceAll(fieldValue, ",", "")
					marketValue, err = strconv.ParseFloat(fieldValue, 64)
					if err != nil {
						log.Println(err)
						log.Fatal(err)
					}
				} else if index == 4 {
					cusip = fieldValue
				} else if index == 5 {
					fieldValue = strings.ReplaceAll(fieldValue, ",", "")
					shares, err = strconv.ParseFloat(fieldValue, 64)
					if err != nil {
						log.Println(err)
						log.Fatal(err)
					}
				}
			})
			obj := model.Stock{
				Date:        timeStamp,
				Fund:        etf,
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
		})

	})

	if len(dict) == 0 {
		return
	}
	DB.DailyETFHoldingDBSave(dict, timeStamp, etf)
	//result := DB.DailyChanges(etf)
	//fmt.Println("Updated daily changes for SPCX", result)
}

//type Stock struct {
//	Date        time.Time
//	Fund        string
//	Company     string
//	Ticker      string
//	Cusip       string
//	Shares      float64
//	MarketValue float64
//	Weight      float64
//	ClosePrice  float64
//}
