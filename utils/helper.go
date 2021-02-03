package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func GetCurrentTime() time.Time {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	return time.Now().In(loc)
}

func ReadWebPage(url string) io.Reader {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	return res.Body
}
