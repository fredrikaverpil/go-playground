package main

import (
	"context"
	"encoding/json"
	"testing"

	"cloud.google.com/go/spanner"
	"gotest.tools/v3/assert"
)

func TestSingersSQL(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "singers.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "singers.sql")
	db := newDB(t, ctx)

	expected := []Artist{
		{SingerID: 1, FirstName: "Marc", LastName: "Richards", Metadata: Metadata{Age: 30, City: "New York"}},
		{SingerID: 2, FirstName: "Catalina", LastName: "Smith", Metadata: Metadata{}},
		{SingerID: 3, FirstName: "Alice", LastName: "Trentor", Metadata: Metadata{}},
		{SingerID: 4, FirstName: "Lea", LastName: "Martin", Metadata: Metadata{}},
		{SingerID: 5, FirstName: "David", LastName: "Lomond", Metadata: Metadata{}},
	}
	got := make([]Artist, 0, len(expected))

	rows, err := db.QueryContext(ctx, `
		SELECT SingerId, FirstName, LastName, Metadata
		FROM Singers
		ORDER BY SingerId
	`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var singerID int64
		var firstName, lastName string
		// go-sql-spanner returns spanner.NullJSON for JSON columns.
		var spannerMetadata spanner.NullJSON
		if err := rows.Scan(&singerID, &firstName, &lastName, &spannerMetadata); err != nil {
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
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration: %v", err)
	}

	assert.DeepEqual(t, got, expected)
}
