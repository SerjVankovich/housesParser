package parser

import (
	"encoding/json"
	"houseParser/utils"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func Run() {
	documents := make(chan *goquery.Document, 100)
	for i := 1; i <= 121; i++ {
		url := "https://dom.mingkh.ru/kareliya/petrozavodsk/?page=" + strconv.Itoa(i)
		time.Sleep(100 * time.Millisecond)
		go GetDocument(url, documents)

	}
	var allHouses []*HouseModel

	for i := 1; i <= 121; i++ {
		allHouses = append(allHouses, Parse(<-documents)...)
	}

	for i := 0; i < len(allHouses); i++ {
		println(allHouses[i].Street, allHouses[i].HouseNumber)
	}

	p, err := json.Marshal(allHouses)
	utils.ProcessError(err)

	f, err := os.Create("output.json")
	utils.ProcessError(err)
	defer f.Close()
	f.Write(p)
}
