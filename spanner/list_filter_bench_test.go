package main

import (
	"context"
	"database/sql"
	"testing"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func BenchmarkListFilter(b *testing.B) {
	ctx := context.Background()
	applySchema(b, ctx, "list_filter.sql")
	client := newClient(b, ctx)
	applySeed(b, ctx, client, "list_filter.sql")
	db := newDB(b, ctx)

	query := `SELECT SongId, Title, Artist, Genre, Year FROM Tracks WHERE Genre = @genre ORDER BY SongId`

	b.Run("spanner", func(b *testing.B) {
		for b.Loop() {
			stmt := spanner.Statement{
				SQL:    query,
				Params: map[string]any{"genre": "Rock"},
			}
			iter := client.Single().Query(ctx, stmt)
			for {
				row, err := iter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					b.Fatalf("read row: %v", err)
				}
				var s Song
				if err := row.Columns(&s.SongId, &s.Title, &s.Artist, &s.Genre, &s.Year); err != nil {
					b.Fatalf("scan columns: %v", err)
				}
			}
			iter.Stop()
		}
	})

	b.Run("database_sql", func(b *testing.B) {
		for b.Loop() {
			rows, err := db.QueryContext(ctx, query, sql.Named("genre", "Rock"))
			if err != nil {
				b.Fatalf("query: %v", err)
			}
			for rows.Next() {
				var s Song
				if err := rows.Scan(&s.SongId, &s.Title, &s.Artist, &s.Genre, &s.Year); err != nil {
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
