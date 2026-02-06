package main

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func BenchmarkNgram(b *testing.B) {
	ctx := context.Background()
	applySchema(b, ctx, "ngram_bench.sql")
	client := newClient(b, ctx)
	applySeed(b, ctx, client, "ngram_bench.sql")

	b.Run("ngram_min_1", func(b *testing.B) {
		query := `
			SELECT Title
			FROM AlbumsNgramMin1
			WHERE SEARCH_NGRAMS(Title_Tokens, "Hatel Kaliphorn")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Hatel Kaliphorn") DESC
			LIMIT 5
		`
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

	b.Run("ngram_min_2", func(b *testing.B) {
		query := `
			SELECT Title
			FROM AlbumsNgramMin2
			WHERE SEARCH_NGRAMS(Title_Tokens, "Hatel Kaliphorn")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Hatel Kaliphorn") DESC
			LIMIT 5
		`
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
}
