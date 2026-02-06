package main

import (
	"context"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"gotest.tools/v3/assert"
)

// TestPhoneticSearch demonstrates Spanner's SOUNDEX-based phonetic search.
// SOUNDEX maps words that sound alike to the same code, enabling searches
// that find results despite different spellings of similar-sounding names.
func TestPhoneticSearch(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "phonetic_search.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "phonetic_search.sql")

	t.Run("soundex codes for similar names", func(t *testing.T) {
		// Verify that SOUNDEX maps similar-sounding names to the same code.
		// "Steven", "Stephen", and "Stefan" should all produce the same soundex.
		stmt := spanner.NewStatement(`
			SELECT FirstName, LastName, FirstNameSoundex
			FROM Artists
			WHERE FirstNameSoundex = LOWER(SOUNDEX("Steven"))
			ORDER BY FirstName
		`)
		type result struct {
			FirstName string
			LastName  string
			Soundex   string
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
			if err := row.Columns(&r.FirstName, &r.LastName, &r.Soundex); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, r)
		}

		var names []string
		for _, r := range results {
			names = append(names, r.FirstName+" "+r.LastName)
			t.Logf("  %s %s -> soundex: %s", r.FirstName, r.LastName, r.Soundex)
		}
		// All three Steven/Stephen/Stefan variants should match.
		assert.DeepEqual(t, names, []string{"Stefan Olsson", "Stephen Stills", "Steven Tyler"})
	})

	t.Run("shawn and sean are phonetically equivalent", func(t *testing.T) {
		// "Shawn" and "Sean" sound the same and should share a soundex code.
		// SOUNDEX preserves the first letter, so both must start with 'S' to match.
		stmt := spanner.NewStatement(`
			SELECT FirstName, LastName, FirstNameSoundex
			FROM Artists
			WHERE FirstNameSoundex = LOWER(SOUNDEX("Shawn"))
			ORDER BY FirstName
		`)
		var names []string
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
			var firstName, lastName, soundex string
			if err := row.Columns(&firstName, &lastName, &soundex); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			names = append(names, firstName+" "+lastName)
			t.Logf("  %s %s -> soundex: %s", firstName, lastName, soundex)
		}
		assert.DeepEqual(t, names, []string{"Sean Lennon", "Shawn Colvin"})
	})

	t.Run("soundex distinguishes different sounds", func(t *testing.T) {
		// "Jon"/"John"/"Johnny" should match each other but NOT "Shawn"/"Sean".
		stmt := spanner.NewStatement(`
			SELECT FirstName, LastName
			FROM Artists
			WHERE FirstNameSoundex = LOWER(SOUNDEX("John"))
			ORDER BY FirstName
		`)
		var names []string
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
			var firstName, lastName string
			if err := row.Columns(&firstName, &lastName); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			names = append(names, firstName+" "+lastName)
		}
		// Jon and John should match. Johnny may or may not depending on SOUNDEX behavior.
		assert.Assert(t, len(names) >= 2, "expected at least 2 results, got %d", len(names))
		t.Logf("  matches for 'John': %v", names)
		// Verify Shawn/Sean are NOT in the results.
		for _, name := range names {
			assert.Assert(t, name != "Shawn Colvin", "Shawn should not match John's soundex")
			assert.Assert(t, name != "Sean Lennon", "Sean should not match John's soundex")
		}
	})
}
