package main

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func BenchmarkFuzzySearch(b *testing.B) {
	ctx := context.Background()
	applySchema(b, ctx, "fuzzy_search.sql")
	client := newClient(b, ctx)
	applySeed(b, ctx, client, "fuzzy_search.sql")
	db := newDB(b, ctx)

	query := `
		SELECT Title
		FROM Albums
		WHERE SEARCH_NGRAMS(Title_Tokens, "Hatel Kaliphorn")
		ORDER BY SCORE_NGRAMS(Title_Tokens, "Hatel Kaliphorn") DESC
		LIMIT 5
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
				var title string
				if err := row.Columns(&title); err != nil {
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
				var title string
				if err := rows.Scan(&title); err != nil {
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
