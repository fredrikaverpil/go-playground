package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"cloud.google.com/go/spanner/spansql"
	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
	"go.einride.tech/spanner-aip/spanfiltering"
	"go.einride.tech/spanner-aip/spanordering"
	"gotest.tools/v3/assert"
)

// listSongsSQL queries Tracks with AIP-160 filtering, AIP-132 ordering, and pagination using database/sql.
func listSongsSQL(ctx context.Context, db *sql.DB, req ListSongsRequest) (*ListSongsResponse, error) {
	declarations := songDeclarations()

	var whereParts []string
	var args []any

	// Parse and transpile filter.
	if req.Filter != "" {
		filter, err := filtering.ParseFilterString(req.Filter, declarations)
		if err != nil {
			return nil, fmt.Errorf("parse filter: %w", err)
		}
		whereExpr, filterParams, err := spanfiltering.TranspileFilter(filter)
		if err != nil {
			return nil, fmt.Errorf("transpile filter: %w", err)
		}
		whereParts = append(whereParts, whereExpr.SQL())
		for k, v := range filterParams {
			args = append(args, sql.Named(k, v))
		}
	}

	// Parse and transpile order_by.
	var orderParts []string
	if req.OrderBy != "" {
		var ob ordering.OrderBy
		if err := ob.UnmarshalString(req.OrderBy); err != nil {
			return nil, fmt.Errorf("parse order_by: %w", err)
		}
		for _, o := range spanordering.TranspileOrderBy(ob) {
			orderParts = append(orderParts, o.SQL())
		}
	}
	// Always include SongId as tiebreaker for deterministic ordering.
	orderParts = append(orderParts, spansql.Order{Expr: spansql.ID("SongId")}.SQL())

	// Determine page size and offset.
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	var offset int64
	if req.PageToken != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.PageToken)
		if err != nil {
			return nil, fmt.Errorf("decode page token: %w", err)
		}
		offset, err = strconv.ParseInt(string(decoded), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse page token offset: %w", err)
		}
	}

	// Build raw SQL.
	query := "SELECT SongId, Title, Artist, Genre, Year FROM Tracks"
	if len(whereParts) > 0 {
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}
	query += " ORDER BY " + strings.Join(orderParts, ", ")
	query += fmt.Sprintf(" LIMIT %d", pageSize+1)
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var songs []Song
	for rows.Next() {
		var s Song
		if err := rows.Scan(&s.SongId, &s.Title, &s.Artist, &s.Genre, &s.Year); err != nil {
			return nil, fmt.Errorf("scan columns: %w", err)
		}
		songs = append(songs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	// If we got more than pageSize results, there's a next page.
	resp := &ListSongsResponse{}
	if len(songs) > int(pageSize) {
		resp.Songs = songs[:pageSize]
		nextOffset := offset + int64(pageSize)
		resp.NextPageToken = base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(nextOffset, 10)))
	} else {
		resp.Songs = songs
	}

	return resp, nil
}

// TestListFilterSQL demonstrates AIP-132 List with AIP-160 filtering using database/sql.
func TestListFilterSQL(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "list_filter.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "list_filter.sql")
	db := newDB(t, ctx)

	t.Run("no filter returns all songs", func(t *testing.T) {
		resp, err := listSongsSQL(ctx, db, ListSongsRequest{})
		assert.NilError(t, err)
		assert.Equal(t, len(resp.Songs), 10)
		assert.Equal(t, resp.NextPageToken, "")
	})

	t.Run("filter by genre", func(t *testing.T) {
		resp, err := listSongsSQL(ctx, db, ListSongsRequest{
			Filter: `Genre = "Rock"`,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(resp.Songs), 4)
		for _, s := range resp.Songs {
			assert.Equal(t, s.Genre, "Rock")
		}
	})

	t.Run("filter with AND", func(t *testing.T) {
		resp, err := listSongsSQL(ctx, db, ListSongsRequest{
			Filter: `Genre = "Rock" AND Year > 1975`,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(resp.Songs), 2)
		for _, s := range resp.Songs {
			assert.Equal(t, s.Genre, "Rock")
			assert.Assert(t, s.Year > 1975, "expected year > 1975, got %d", s.Year)
		}
	})

	t.Run("filter with OR", func(t *testing.T) {
		resp, err := listSongsSQL(ctx, db, ListSongsRequest{
			Filter: `Genre = "Rock" OR Genre = "Jazz"`,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(resp.Songs), 7)
		for _, s := range resp.Songs {
			assert.Assert(t, s.Genre == "Rock" || s.Genre == "Jazz",
				"expected Rock or Jazz, got %s", s.Genre)
		}
	})

	t.Run("order by year ascending", func(t *testing.T) {
		resp, err := listSongsSQL(ctx, db, ListSongsRequest{
			OrderBy: "Year",
		})
		assert.NilError(t, err)
		assert.Assert(t, len(resp.Songs) > 0)
		for i := 1; i < len(resp.Songs); i++ {
			assert.Assert(t, resp.Songs[i].Year >= resp.Songs[i-1].Year,
				"expected ascending order: %d >= %d", resp.Songs[i].Year, resp.Songs[i-1].Year)
		}
	})

	t.Run("order by year descending", func(t *testing.T) {
		resp, err := listSongsSQL(ctx, db, ListSongsRequest{
			OrderBy: "Year desc",
		})
		assert.NilError(t, err)
		assert.Assert(t, len(resp.Songs) > 0)
		for i := 1; i < len(resp.Songs); i++ {
			assert.Assert(t, resp.Songs[i].Year <= resp.Songs[i-1].Year,
				"expected descending order: %d <= %d", resp.Songs[i].Year, resp.Songs[i-1].Year)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		var allSongs []Song
		var pageToken string
		pages := 0
		for {
			resp, err := listSongsSQL(ctx, db, ListSongsRequest{
				PageSize:  3,
				PageToken: pageToken,
				OrderBy:   "SongId",
			})
			assert.NilError(t, err)
			allSongs = append(allSongs, resp.Songs...)
			pages++
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		assert.Equal(t, pages, 4)
		assert.Equal(t, len(allSongs), 10)
		for i, s := range allSongs {
			assert.Equal(t, s.SongId, int64(i+1))
		}
	})

	t.Run("invalid filter returns error", func(t *testing.T) {
		_, err := listSongsSQL(ctx, db, ListSongsRequest{
			Filter: `Genre = 123`,
		})
		assert.Assert(t, err != nil, "expected error for type mismatch filter")
	})
}
