package model

type Department struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Transactions []TransactionHolder
}
