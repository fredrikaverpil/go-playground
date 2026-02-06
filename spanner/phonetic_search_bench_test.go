package main

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func BenchmarkPhoneticSearch(b *testing.B) {
	ctx := context.Background()
	applySchema(b, ctx, "phonetic_search.sql")
	client := newClient(b, ctx)
	applySeed(b, ctx, client, "phonetic_search.sql")
	db := newDB(b, ctx)

	query := `
		SELECT FirstName, LastName, FirstNameSoundex
		FROM Artists
		WHERE FirstNameSoundex = LOWER(SOUNDEX("Steven"))
		ORDER BY FirstName
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
				var firstName, lastName, soundex string
				if err := row.Columns(&firstName, &lastName, &soundex); err != nil {
					b.Fatalf("scan columns: %v", err)
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
				var firstName, lastName, soundex string
				if err := rows.Scan(&firstName, &lastName, &soundex); err != nil {
					b.Fatalf("scan columns: %v", err)
				}
			}
			_ = rows.Close()
			if err := rows.Err(); err != nil {
				b.Fatalf("rows iteration: %v", err)
			}
		}
	})
}
