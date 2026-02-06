package main

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestFullTextSearchSQL(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "fulltext_search.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "fulltext_search.sql")
	db := newDB(t, ctx)

	t.Run("single word search in title", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT SongId, Title
			FROM Songs
			WHERE SEARCH(Title_Tokens, 'ocean')
			ORDER BY SongId
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		var results []string
		for rows.Next() {
			var songID int64
			var title string
			if err := rows.Scan(&songID, &title); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, title)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}
		assert.DeepEqual(t, results, []string{"Ocean Drive", "Ocean Waves"})
	})

	t.Run("search in description", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT Title
			FROM Songs
			WHERE SEARCH(Description_Tokens, 'rain')
			ORDER BY Title
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
		assert.DeepEqual(t, results, []string{"City Rain", "Forest Rain"})
	})

	t.Run("boolean OR search", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT Title
			FROM Songs
			WHERE SEARCH(Description_Tokens, 'desert OR mountain')
			ORDER BY Title
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
		assert.DeepEqual(t, results, []string{"Desert Wind", "Mountain Echo"})
	})

	t.Run("scoring and ranking", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT Title, SCORE(Description_Tokens, 'ocean') AS relevance
			FROM Songs
			WHERE SEARCH(Description_Tokens, 'ocean')
			ORDER BY relevance DESC
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		type result struct {
			Title     string
			Relevance float64
		}
		var results []result
		for rows.Next() {
			var r result
			if err := rows.Scan(&r.Title, &r.Relevance); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, r)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}
		assert.Assert(t, len(results) >= 2, "expected at least 2 results, got %d", len(results))
		for _, r := range results {
			t.Logf("  %s (score: %f)", r.Title, r.Relevance)
			assert.Assert(t, r.Relevance > 0, "expected positive score for %s", r.Title)
		}
	})
}
