package transactions

import (
	"fmt"
	"regexp"
	"time"
)

type TX struct {
	Hash          string    `bson:"hash" json:"hash"`
	From          string    `bson:"from" json:"from"`
	To            string    `bson:"to" json:"to"`
	BlockNumber   int64     `bson:"blockNumber" json:"blockNumber"`
	Value         float64   `bson:"valueEth" json:"value"`
	Fee           float64   `bson:"feeEth" json:"fee"`
	Confirmations int64     `bson:"confirmations" json:"confirmations"`
	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
}

type TXFilter struct {
	Page     int64     `json:"page"`
	PageSize int64     `json:"page_size"`
	Hash     string    `json:"hash"`
	From     string    `json:"from"`
	To       string    `json:"to"`
	BlockNum int64     `json:"block_num"`
	DateFrom time.Time `json:"date_from"`
	DateTo   time.Time `json:"date_to"`
}

func (f *TXFilter) Validate() error {
	if f.Page > 0 {
		f.Page -= 1 // to start counting from the first page
	}
	if f.PageSize > 1000 {
		return fmt.Errorf("page_size max value = 1000")
	}

	var isHash = regexp.MustCompile(`^[A-Fa-f0-9]{64}$`).MatchString
	var isAddress = regexp.MustCompile(`^[A-Fa-f0-9]{40}$`).MatchString

	if len(f.Hash) != 0 && !isHash(f.Hash) {
		return fmt.Errorf("invalid hash")
	}
	if len(f.From) != 0 && !isAddress(f.From) {
		return fmt.Errorf("invalid 'from' address")
	}
	if len(f.To) != 0 && !isAddress(f.To) {
		return fmt.Errorf("invalid 'to' address")
	}

	return nil
}
