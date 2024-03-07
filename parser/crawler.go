package parser

import (
	"fmt"
	"houseParser/utils"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func GetDocument(url string, documents chan *goquery.Document) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error while retrieving site", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.8")

	res, err := client.Do(req)
	utils.ProcessError(err)

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	utils.ProcessError(err)
	println("Fetched page: " + url)
	documents <- doc
}
