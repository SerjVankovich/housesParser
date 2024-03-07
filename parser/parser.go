package parser

import (
	"houseParser/utils"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Parse(document *goquery.Document) []*HouseModel {
	var houses []*HouseModel
	document.Find(".table.table-condensed.table-hover.table-striped").Each(func(i int, table *goquery.Selection) {
		tbody := table.Find("tbody")
		tbody.Find("tr").Each(func(i int, tr *goquery.Selection) {
			tds := tr.Find("td").Map(func(i int, s *goquery.Selection) string { return s.Text() })
			address := tds[1]
			levels, err := strconv.Atoi(strings.Trim(tds[3], " "))
			if err != nil {
				levels = 0
			}
			area, err := strconv.ParseFloat(strings.Trim(tds[4], " "), 64)
			if err != nil {
				area = 0
			}
			street, houseNumber := utils.ProcessAddress(address)
			houses = append(houses, &HouseModel{street, levels, area, houseNumber})
		})
	})
	println("Houses fetched: " + strconv.Itoa(len(houses)))

	return houses
}
