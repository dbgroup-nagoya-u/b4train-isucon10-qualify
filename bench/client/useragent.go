package client

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

var browserList []string = []string{
	"ISUCON Nickel",
	"ISUCON Icetanuki",
	"ISUCON Web Browser",
	"ISUCON Explorer",
	"ISUCON Edge",
}

var suffixList []string = []string{
	" mobile",
	" bottle",
	" alpha",
	" beta",
	"",
}

func GenerateUserAgent() string {
	browser := browserList[rand.Intn(len(browserList))]
	suffix := suffixList[rand.Intn(len(suffixList))]
	var uuidStr string
	u, err := uuid.NewRandom()
	if err != nil {
		uuidStr = "00000000-0000-0000-0000-000000000000"
	} else {
		uuidStr = u.String()
	}

	return fmt.Sprintf("%v%v-%v", browser, suffix, uuidStr)
}
