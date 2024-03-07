package utils

import (
	"log"
	"strings"
)

func ProcessError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ProcessAddress(address string) (string, string) {
	splittedAddress := strings.Split(address, ",")
	street := strings.Trim(splittedAddress[1], " ")
	houseNumber := strings.Trim(splittedAddress[2], " ")
	return street, houseNumber
}
