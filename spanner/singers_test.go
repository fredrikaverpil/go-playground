package main

// Metadata holds optional JSON metadata for a singer.
type Metadata struct {
	Age  int    `json:"age"`
	City string `json:"city"`
}

// Artist represents a row in the Singers table.
type Artist struct {
	SingerID  int64
	FirstName string
	LastName  string
	Metadata  Metadata
}
