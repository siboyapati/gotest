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
	"sync"
	"time"
)

const (
	CONNECTIONSTRING = ""
	DB     = "sample_airbnb"
	ISSUES = "col_issues"
	ARKDB   = "etf_db"
	TRANSDB = "etf_activity"
	TEST = "test"
)


var mongoOnce sync.Once
var clientInstance *mongo.Client
var clientInstanceError error

func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		clientOptions := options.Client().ApplyURI(CONNECTIONSTRING)
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			clientInstanceError = err
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			clientInstanceError = err
		}
		clientInstance = client
	})
	return clientInstance, clientInstanceError
}

// Do not insert if there is are no changes in the ETF
func UpdateDB(etf string, dict map[string]model.Stock) bool {
	client, err := GetMongoClient()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	collection := client.Database(ARKDB).Collection(etf)
	result := model.ArkResultResponse{}
	opts := options.FindOne().SetSort(bson.D{{"timestamp", -1}})
	err = collection.FindOne(context.TODO(), bson.D{}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return true
		}
		fmt.Println("HasRecord err =", err)
		log.Fatal(err)
	}

	filter := bson.D{{"_id", result.ID}}
	update := bson.D{{"$set",
		bson.D{
			{"stocks", dict},
		},
	}}

	res, err := client.Database(ARKDB).Collection(etf).UpdateOne(
		context.TODO(),
		filter,
		update,
	)
	if err != nil {
		fmt.Println(err, etf)
	}

	fmt.Println(res)
	//if len(result.Stocks) != len(dict){
	//	return true
	//}
	//for ticker, prevStock := range result.Stocks{
	//	//fmt.Println(ticker, prevStock)
	//	if curStock, ok := dict[ticker]; !ok{
	//		return true
	//	} else {
	//		if curStock.Shares != prevStock.Shares{
	//			return true
	//		}
	//	}
	//}
	DBResult := DailyChanges(etf)
	fmt.Println("Updated daily changes for =", etf, DBResult)
	return true
}

// Do not insert if there is are no changes in the ETF
func ShouldInsert(etf string, dict map[string]model.Stock) bool {
	client, err := GetMongoClient()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	collection := client.Database(ARKDB).Collection(etf)
	result := model.Ark{}
	opts := options.FindOne().SetSort(bson.D{{"timestamp", -1}})
	err = collection.FindOne(context.TODO(), bson.D{}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return true
		}
		fmt.Println("HasRecord err =", err)
		log.Fatal(err)
	}
	if len(result.Stocks) != len(dict) {
		return true
	}
	for ticker, prevStock := range result.Stocks {
		//fmt.Println(ticker, prevStock)
		if curStock, ok := dict[ticker]; !ok {
			return true
		} else {
			if curStock.Shares != prevStock.Shares {
				return true
			}
		}
	}
	return false
}

func DailyETFHoldingDBSave(dict map[string]model.Stock, timeStamp time.Time, etf string) {
	arkObj := model.Ark{
		Stocks:    dict,
		Date:      utils.GetCurrentTime(),
		TimeStamp: timeStamp,
	}

	//if !shouldInsert(etf, dict) {
	//	fmt.Println("There is no change in the ETF && we are not inserting into the database.", etf)
	//	return
	//}
	client, err := GetMongoClient()
	if err != nil {
		fmt.Println(err)
	}

	collection := client.Database(ARKDB).Collection(etf)
	result, err := collection.InsertOne(context.TODO(), arkObj)

	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	fmt.Println("Insert success for:", result, etf)
	DBResult := DailyChanges(etf)
	fmt.Println("Updated daily changes for =", etf, DBResult)
}

//http://ishares.com//us/products/308878/fund/1467271812596.ajax?fileType=csv&fileName=IDNA_holdings&dataType=fund
//https://www.globalxetfs.com/funds/gnom/?download_full_holdings=true
//https://www.renaissancecapital.com//IPO-Investing/US-IPO-ETF-Holdings-Excel-Download

//http://ishares.com/us/products/308878/fund/1467271812596.ajax?fileType=csv&fileName=IDNA_holdings&dataType=fund
//http://ishares.com/us/products/239705/ishares-phlx-semiconductor-etf/1467271812596.ajax?fileType=csv&fileName=SOXX_holdings&dataType=fund

//Genomics
//https://etfmg.com/funds/germ/
//https://principaletfs.com/etf/btec#holdings
