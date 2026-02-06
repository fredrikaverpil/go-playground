package main

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"gotest.tools/v3/assert"
)

// TestFullTextSearch demonstrates Spanner's full-text search capabilities
// using TOKENIZE_FULLTEXT, SEARCH INDEX, SEARCH(), and SCORE().
func TestFullTextSearch(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "fulltext_search.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "fulltext_search.sql")

	t.Run("single word search in title", func(t *testing.T) {
		// SEARCH(column_tokens, query) returns true if tokens match the query.
		stmt := spanner.NewStatement(`
			SELECT SongId, Title
			FROM Songs
			WHERE SEARCH(Title_Tokens, 'ocean')
			ORDER BY SongId
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
			var songID int64
			var title string
			if err := row.Columns(&songID, &title); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, title)
		}
		// Should find both ocean-related titles.
		assert.DeepEqual(t, results, []string{"Ocean Drive", "Ocean Waves"})
	})

	t.Run("search in description", func(t *testing.T) {
		// Search across description text for the word "rain".
		stmt := spanner.NewStatement(`
			SELECT Title
			FROM Songs
			WHERE SEARCH(Description_Tokens, 'rain')
			ORDER BY Title
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
		// "City Rain" and "Forest Rain" both mention rain in their descriptions.
		assert.DeepEqual(t, results, []string{"City Rain", "Forest Rain"})
	})

	t.Run("boolean OR search", func(t *testing.T) {
		// Use OR to match songs about either desert or mountain topics.
		stmt := spanner.NewStatement(`
			SELECT Title
			FROM Songs
			WHERE SEARCH(Description_Tokens, 'desert OR mountain')
			ORDER BY Title
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
		assert.DeepEqual(t, results, []string{"Desert Wind", "Mountain Echo"})
	})

	t.Run("scoring and ranking", func(t *testing.T) {
		// SCORE() returns a relevance score for ranking results.
		// Songs with "ocean" in both title and description should score higher.
		stmt := spanner.NewStatement(`
			SELECT Title, SCORE(Description_Tokens, 'ocean') AS relevance
			FROM Songs
			WHERE SEARCH(Description_Tokens, 'ocean')
			ORDER BY relevance DESC
		`)
		type result struct {
			Title     string
			Relevance float64
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
			if err := row.Columns(&r.Title, &r.Relevance); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, r)
		}
		// Both ocean songs should appear, with scores > 0.
		assert.Assert(t, len(results) >= 2, "expected at least 2 results, got %d", len(results))
		for _, r := range results {
			t.Logf("  %s (score: %f)", r.Title, r.Relevance)
			assert.Assert(t, r.Relevance > 0, "expected positive score for %s", r.Title)
		}
	})
}
