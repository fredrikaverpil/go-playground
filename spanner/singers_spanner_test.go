package main

import (
	"context"
	"encoding/json"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"gotest.tools/v3/assert"
)

type Metadata struct {
	Age  int    `json:"age"`
	City string `json:"city"`
}

type Artist struct {
	SingerID  int64
	FirstName string
	LastName  string
	Metadata  Metadata
}

func TestSingersInsert(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "singers.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "singers.sql")

	expected := []Artist{
		{SingerID: 1, FirstName: "Marc", LastName: "Richards", Metadata: Metadata{Age: 30, City: "New York"}},
		{SingerID: 2, FirstName: "Catalina", LastName: "Smith", Metadata: Metadata{}},
		{SingerID: 3, FirstName: "Alice", LastName: "Trentor", Metadata: Metadata{}},
		{SingerID: 4, FirstName: "Lea", LastName: "Martin", Metadata: Metadata{}},
		{SingerID: 5, FirstName: "David", LastName: "Lomond", Metadata: Metadata{}},
	}
	got := make([]Artist, 0, len(expected))

	iter := client.Single().Query(ctx, spanner.NewStatement(`
		SELECT SingerId, FirstName, LastName, Metadata
		FROM Singers
		ORDER BY SingerId
	`))
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatalf("read row: %v", err)
		}
		var singerID int64
		var firstName, lastName string
		var spannerMetadata spanner.NullJSON
		if err := row.Columns(&singerID, &firstName, &lastName, &spannerMetadata); err != nil {
			t.Fatalf("scan columns: %v", err)
		}
		var metadata Metadata
		if spannerMetadata.Valid {
			jsonBytes, err := json.Marshal(spannerMetadata.Value)
			if err != nil {
				t.Fatalf("marshal JSON value: %v", err)
			}
			if err := json.Unmarshal(jsonBytes, &metadata); err != nil {
				t.Fatalf("unmarshal JSON: %v", err)
			}
		}
		got = append(got, Artist{SingerID: singerID, FirstName: firstName, LastName: lastName, Metadata: metadata})
	}

	assert.DeepEqual(t, got, expected)
}
