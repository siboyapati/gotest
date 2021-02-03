package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Stock struct {
	Date        time.Time
	Fund        string
	Company     string
	Ticker      string
	Cusip       string
	Shares      float64
	MarketValue float64
	Weight      float64
	ClosePrice  float64
}

type Ark struct {
	Stocks    map[string]Stock
	Date      time.Time
	TimeStamp time.Time
}


type ArkResultResponse struct {
	Stocks    map[string]Stock
	Date      time.Time
	TimeStamp time.Time
    ID primitive.ObjectID `bson:"_id, omitempty"`
}

type ArkList struct {
	list []Ark
}

type ArkTransaction struct {
	//Date        time.Time
	TimeStamp   time.Time
	Fund        string
	Company     string
	Ticker      string
	Cusip       string
	TotalShares float64
	MarketValue float64
	Shares      float64
    TransactionCost float64
	ClosePrice  float64
	Buy         bool
	NewPosition bool
	ShareChangePercentage float64
}


type TransactionList struct {
	DailyTransactions []ArkTransaction
	Date             time.Time
	TimeStamp        time.Time
}
