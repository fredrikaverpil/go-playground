package main

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

// ngramConfigs defines the table/sub-benchmark pairs for the n-gram size comparison.
var ngramConfigs = []struct {
	name  string
	table string
}{
	{name: "ngram_min_1", table: "AlbumsNgramMin1"},
	{name: "ngram_min_2", table: "AlbumsNgramMin2"},
	{name: "ngram_min_3", table: "AlbumsNgramMin3"},
}

func BenchmarkNgram(b *testing.B) {
	client := setupNgramBench(b)
	for _, cfg := range ngramConfigs {
		b.Run(cfg.name, ngramBenchFunc(client, cfg.table))
	}
}

func BenchmarkNgramReversed(b *testing.B) {
	client := setupNgramBench(b)
	for i := len(ngramConfigs) - 1; i >= 0; i-- {
		cfg := ngramConfigs[i]
		b.Run(cfg.name, ngramBenchFunc(client, cfg.table))
	}
}

// setupNgramBench applies schema, seeds data, and warms up all tables.
func setupNgramBench(b *testing.B) *spanner.Client {
	b.Helper()
	ctx := context.Background()
	applySchema(b, ctx, "ngram_bench.sql")
	client := newClient(b, ctx)
	applySeed(b, ctx, client, "ngram_bench.sql")

	// Warm up all tables so the emulator has compiled query plans and loaded
	// index data before the timed benchmarks start.
	for _, cfg := range ngramConfigs {
		query := fmt.Sprintf(`
			SELECT Title
			FROM %s
			WHERE SEARCH_NGRAMS(Title_Tokens, "Hatel Kaliphorn")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Hatel Kaliphorn") DESC
			LIMIT 5
		`, cfg.table)
		for range 10 {
			iter := client.Single().Query(ctx, spanner.NewStatement(query))
			for {
				row, err := iter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					b.Fatalf("warm-up query on %s: %v", cfg.table, err)
				}
				var title string
				if err := row.Columns(&title); err != nil {
					b.Fatalf("warm-up scan on %s: %v", cfg.table, err)
				}
			}
			iter.Stop()
		}
	}
	return client
}

// ngramBenchFunc returns a benchmark function that queries the given table.
func ngramBenchFunc(client *spanner.Client, table string) func(*testing.B) {
	return func(b *testing.B) {
		ctx := context.Background()
		query := fmt.Sprintf(`
			SELECT Title
			FROM %s
			WHERE SEARCH_NGRAMS(Title_Tokens, "Hatel Kaliphorn")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Hatel Kaliphorn") DESC
			LIMIT 5
		`, table)
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
	}
}
