package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func BenchmarkSingers(b *testing.B) {
	ctx := context.Background()
	applySchema(b, ctx, "singers.sql")
	client := newClient(b, ctx)
	applySeed(b, ctx, client, "singers.sql")
	db := newDB(b, ctx)

	query := `
		SELECT SingerId, FirstName, LastName, Metadata
		FROM Singers
		ORDER BY SingerId
	`

	b.Run("spanner", func(b *testing.B) {
		for b.Loop() {
			iter := client.Single().Query(ctx, spanner.NewStatement(query))
			for {
				row, err := iter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					b.Fatalf("read row: %v", err)
				}
				var singerID int64
				var firstName, lastName string
				var spannerMetadata spanner.NullJSON
				if err := row.Columns(&singerID, &firstName, &lastName, &spannerMetadata); err != nil {
					b.Fatalf("scan columns: %v", err)
				}
				var metadata Metadata
				if spannerMetadata.Valid {
					jsonBytes, _ := json.Marshal(spannerMetadata.Value)
					_ = json.Unmarshal(jsonBytes, &metadata)
				}
			}
			iter.Stop()
		}
	})

	b.Run("database_sql", func(b *testing.B) {
		for b.Loop() {
			rows, err := db.QueryContext(ctx, query)
			if err != nil {
				b.Fatalf("query: %v", err)
			}
			for rows.Next() {
				var singerID int64
				var firstName, lastName string
				var rawJSON sql.NullString
				if err := rows.Scan(&singerID, &firstName, &lastName, &rawJSON); err != nil {
					b.Fatalf("scan columns: %v", err)
				}
				var metadata Metadata
				if rawJSON.Valid {
					_ = json.Unmarshal([]byte(rawJSON.String), &metadata)
				}
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				b.Fatalf("rows iteration: %v", err)
			}
		}
	})
}
