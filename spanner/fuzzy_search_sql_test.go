package main

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestFuzzySearchSQL(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "fuzzy_search.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "fuzzy_search.sql")
	db := newDB(t, ctx)

	t.Run("misspelled query finds correct result", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT Title
			FROM Albums
			WHERE SEARCH_NGRAMS(Title_Tokens, "Hatel Kaliphorn")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Hatel Kaliphorn") DESC
			LIMIT 5
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		var results []string
		for rows.Next() {
			var title string
			if err := rows.Scan(&title); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, title)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}
		assert.Assert(t, len(results) > 0, "expected at least one result")
		assert.Equal(t, results[0], "Hotel California")
		t.Logf("  top result for 'Hatel Kaliphorn': %s", results[0])
	})

	t.Run("scoring ranks closer matches higher", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT Title, SCORE_NGRAMS(Title_Tokens, "Abey Road") AS score
			FROM Albums
			WHERE SEARCH_NGRAMS(Title_Tokens, "Abey Road")
			ORDER BY score DESC
			LIMIT 5
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		type result struct {
			Title string
			Score float64
		}
		var results []result
		for rows.Next() {
			var r result
			if err := rows.Scan(&r.Title, &r.Score); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, r)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}
		assert.Assert(t, len(results) > 0, "expected at least one result")
		t.Log("  results for 'Abey Road':")
		for _, r := range results {
			t.Logf("    %s (score: %f)", r.Title, r.Score)
		}
		assert.Equal(t, results[0].Title, "Abbey Road")
	})

	t.Run("partial word match", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT Title
			FROM Albums
			WHERE SEARCH_NGRAMS(Title_Tokens, "Nevermi")
			ORDER BY SCORE_NGRAMS(Title_Tokens, "Nevermi") DESC
			LIMIT 3
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		var results []string
		for rows.Next() {
			var title string
			if err := rows.Scan(&title); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, title)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}
		assert.Assert(t, len(results) > 0, "expected at least one result")
		assert.Equal(t, results[0], "Nevermind")
		t.Logf("  top result for 'Nevermi': %s", results[0])
	})
}
