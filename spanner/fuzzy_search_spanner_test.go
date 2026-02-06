package main

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"gotest.tools/v3/assert"
)

// TestFuzzySearch demonstrates Spanner's n-gram-based approximate search
// using TOKENIZE_SUBSTRING, SEARCH_NGRAMS(), and SCORE_NGRAMS().
// This enables finding results even when the search query contains typos.
func TestFuzzySearch(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "fuzzy_search.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "fuzzy_search.sql")

	t.Run("misspelled query finds correct result", func(t *testing.T) {
		// SEARCH_NGRAMS finds candidates sharing n-grams with the query.
		// "Hatel Kaliphorn" shares enough n-grams with "Hotel California" to match.
		stmt := spanner.NewStatement(`
			SELECT Title
			FROM Albums
			WHERE SEARCH_NGRAMS(Title_Tokens, "Hatel Kaliphorn")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Hatel Kaliphorn") DESC
			LIMIT 5
		`)
		var results []string
		iter := client.Single().Query(ctx, stmt)
		defer iter.Stop()
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("read row: %v", err)
			}
			var title string
			if err := row.Columns(&title); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, title)
		}
		// "Hotel California" should be the top result.
		assert.Assert(t, len(results) > 0, "expected at least one result")
		assert.Equal(t, results[0], "Hotel California")
		t.Logf("  top result for 'Hatel Kaliphorn': %s", results[0])
	})

	t.Run("scoring ranks closer matches higher", func(t *testing.T) {
		// SCORE_NGRAMS uses Jaccard similarity:
		//   shared_ngrams / (total_ngrams_index + total_ngrams_query - shared_ngrams)
		// A closer match produces a higher score.
		stmt := spanner.NewStatement(`
			SELECT Title, SCORE_NGRAMS(Title_Tokens, "Abey Road") AS score
			FROM Albums
			WHERE SEARCH_NGRAMS(Title_Tokens, "Abey Road")
			ORDER BY score DESC
			LIMIT 5
		`)
		type result struct {
			Title string
			Score float64
		}
		var results []result
		iter := client.Single().Query(ctx, stmt)
		defer iter.Stop()
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("read row: %v", err)
			}
			var r result
			if err := row.Columns(&r.Title, &r.Score); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, r)
		}
		assert.Assert(t, len(results) > 0, "expected at least one result")
		t.Log("  results for 'Abey Road':")
		for _, r := range results {
			t.Logf("    %s (score: %f)", r.Title, r.Score)
		}
		// "Abbey Road" should be the top match.
		assert.Equal(t, results[0].Title, "Abbey Road")
	})

	t.Run("partial word match", func(t *testing.T) {
		// N-gram tokenization also enables partial/prefix matching.
		stmt := spanner.NewStatement(`
			SELECT Title
			FROM Albums
			WHERE SEARCH_NGRAMS(Title_Tokens, "Nevermi")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Nevermi") DESC
			LIMIT 3
		`)
		var results []string
		iter := client.Single().Query(ctx, stmt)
		defer iter.Stop()
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("read row: %v", err)
			}
			var title string
			if err := row.Columns(&title); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, title)
		}
		assert.Assert(t, len(results) > 0, "expected at least one result")
		assert.Equal(t, results[0], "Nevermind")
		t.Logf("  top result for 'Nevermi': %s", results[0])
	})
}
