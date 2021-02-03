package DB

import (
	"context"
	"fmt"
	"github.com/siboyapati/arketf/model"
	"github.com/siboyapati/arketf/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sort"
	"time"
)

func HasRecord(date time.Time, etf string) bool {
	client, err := GetMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(ARKDB).Collection(etf)
	result := model.Ark{}

	opts := options.FindOne().SetSort(bson.D{{"timestamp", -1}})
	err = collection.FindOne(context.TODO(), bson.D{}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false
		}
		fmt.Println("HasRecord err =", err)
		log.Fatal(err)
	}
	return DateEqual(result.TimeStamp, date)
}

func DateEqual(date1, date2 time.Time) bool {
	diff := date2.Sub(date1)
	if diff > 0 {
		return false
	}
	return true
}

func HasTransactionRecord(date time.Time, etf string) bool {
	client, err := GetMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(TRANSDB).Collection(etf)
	result := model.ArkTransaction{}
	opts := options.FindOne().SetSort(bson.D{{"timestamp", -1}})
	err = collection.FindOne(context.TODO(), bson.D{}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false
		}
		fmt.Println("HasRecord err =", err)
		log.Fatal(err)
	}
	return DateEqual(result.TimeStamp, date)
}

func DailyChanges(etf string) bool {
	client, err := GetMongoClient()
	var result []model.Ark
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(ARKDB).Collection(etf)

	opts := options.Find()
	opts.SetSort(bson.D{{"timestamp", -1}})
	opts.SetLimit(2)
	cur, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false
		}
		fmt.Println("HasRecord err =", err)
		log.Fatal(err)
	}

	if err = cur.All(context.Background(), &result); err != nil {
		fmt.Println("HasRecord err / cursor=", err)
		log.Fatal(err)
	}
	var timeStamp time.Time
	if len(result) == 2 {
		curMap := result[0].Stocks
		prevMap := result[1].Stocks

		if HasTransactionRecord(result[0].TimeStamp, etf) {
			fmt.Println("Record already exists #", etf, utils.GetCurrentTime())
			return false
		}

		var transactionArray []model.ArkTransaction
		for key, cur := range curMap {
			timeStamp = cur.Date
			transaction := model.ArkTransaction{
				TimeStamp:   cur.Date,
				Fund:        cur.Fund,
				Company:     cur.Company,
				Ticker:      cur.Ticker,
				Cusip:       cur.Cusip,
				TotalShares: cur.Shares,
				MarketValue: cur.MarketValue,
				NewPosition: false,
			}
			if prev, ok := prevMap[key]; !ok {
				transaction.Buy = true
				transaction.NewPosition = true
				transaction.Shares = cur.Shares
				transaction.TransactionCost = cur.MarketValue
				transaction.ClosePrice = cur.MarketValue / cur.Shares
				transaction.ShareChangePercentage = 100.0

				//fmt.Println(transaction)
			} else {
				if prev.Shares == cur.Shares {
					continue
				}
				//var closePrice float64
				//if cur.ClosePrice == 0 {
				//	closePrice = cur.MarketValue / cur.Shares
				//} else {
				//	closePrice = cur.ClosePrice
				//}
				shareChangePercentage := 100.0
				if prev.Shares != 0 {
					shareChangePercentage = ((cur.Shares - prev.Shares) / prev.Shares) * 100
				}

				closePrice := cur.MarketValue / cur.Shares
				transaction.ClosePrice = cur.MarketValue / cur.Shares
				transaction.Buy = cur.Shares > prev.Shares
				transaction.Shares = cur.Shares - prev.Shares
				transaction.TransactionCost = (cur.Shares - prev.Shares) * closePrice
				transaction.ShareChangePercentage = shareChangePercentage
			}
			transactionArray = append(transactionArray, transaction)
		}
		//fmt.Println("transactionArray unsorted=", transactionArray)
		sort.Slice(transactionArray, func(i, j int) bool {
			return transactionArray[i].TransactionCost > transactionArray[j].TransactionCost
		})

		if len(transactionArray) == 0 {
			fmt.Println("No difference in the ETF", etf)
			return false
		}

		arkObj := model.TransactionList{
			DailyTransactions: transactionArray,
			Date:              utils.GetCurrentTime(),
			TimeStamp:         timeStamp,
		}

		collection := client.Database(TRANSDB).Collection(etf)
		res, err := collection.InsertOne(context.TODO(), arkObj)
		if err != nil {
			log.Fatal(err)
			fmt.Println(err)
		}
		fmt.Println(res)
		//fmt.Println("transactionArray sorted=", transactionArray)
		return true
	}
	return false
}
