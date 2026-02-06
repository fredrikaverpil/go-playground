package main

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPhoneticSearchSQL(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "phonetic_search.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "phonetic_search.sql")
	db := newDB(t, ctx)

	t.Run("soundex codes for similar names", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT FirstName, LastName, FirstNameSoundex
			FROM Artists
			WHERE FirstNameSoundex = LOWER(SOUNDEX("Steven"))
			ORDER BY FirstName
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		type result struct {
			FirstName string
			LastName  string
			Soundex   string
		}
		var results []result
		for rows.Next() {
			var r result
			if err := rows.Scan(&r.FirstName, &r.LastName, &r.Soundex); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			results = append(results, r)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}

		var names []string
		for _, r := range results {
			names = append(names, r.FirstName+" "+r.LastName)
			t.Logf("  %s %s -> soundex: %s", r.FirstName, r.LastName, r.Soundex)
		}
		assert.DeepEqual(t, names, []string{"Stefan Olsson", "Stephen Stills", "Steven Tyler"})
	})

	t.Run("shawn and sean are phonetically equivalent", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT FirstName, LastName, FirstNameSoundex
			FROM Artists
			WHERE FirstNameSoundex = LOWER(SOUNDEX("Shawn"))
			ORDER BY FirstName
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		var names []string
		for rows.Next() {
			var firstName, lastName, soundex string
			if err := rows.Scan(&firstName, &lastName, &soundex); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			names = append(names, firstName+" "+lastName)
			t.Logf("  %s %s -> soundex: %s", firstName, lastName, soundex)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}
		assert.DeepEqual(t, names, []string{"Sean Lennon", "Shawn Colvin"})
	})

	t.Run("soundex distinguishes different sounds", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `
			SELECT FirstName, LastName
			FROM Artists
			WHERE FirstNameSoundex = LOWER(SOUNDEX("John"))
			ORDER BY FirstName
		`)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer func() { _ = rows.Close() }()
		var names []string
		for rows.Next() {
			var firstName, lastName string
			if err := rows.Scan(&firstName, &lastName); err != nil {
				t.Fatalf("scan columns: %v", err)
			}
			names = append(names, firstName+" "+lastName)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows iteration: %v", err)
		}
		assert.Assert(t, len(names) >= 2, "expected at least 2 results, got %d", len(names))
		t.Logf("  matches for 'John': %v", names)
		for _, name := range names {
			assert.Assert(t, name != "Shawn Colvin", "Shawn should not match John's soundex")
			assert.Assert(t, name != "Sean Lennon", "Sean should not match John's soundex")
		}
	})
}
