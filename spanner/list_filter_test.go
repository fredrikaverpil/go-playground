package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"maps"
	"strconv"
	"testing"

	"cloud.google.com/go/spanner"
	"cloud.google.com/go/spanner/spansql"
	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
	"go.einride.tech/spanner-aip/spanfiltering"
	"go.einride.tech/spanner-aip/spanordering"
	"google.golang.org/api/iterator"
	"gotest.tools/v3/assert"
)

// Song represents a row in the Tracks table.
type Song struct {
	SongId int64
	Title  string
	Artist string
	Genre  string
	Year   int64
}

// ListSongsRequest mirrors an AIP-132 List request.
type ListSongsRequest struct {
	Filter    string
	OrderBy   string
	PageSize  int32
	PageToken string
}

// ListSongsResponse mirrors an AIP-132 List response.
type ListSongsResponse struct {
	Songs         []Song
	NextPageToken string
}

// songDeclarations declares filterable fields for the Tracks table.
func songDeclarations() *filtering.Declarations {
	declarations, err := filtering.NewDeclarations(
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent("Genre", filtering.TypeString),
		filtering.DeclareIdent("Year", filtering.TypeInt),
		filtering.DeclareIdent("Artist", filtering.TypeString),
		filtering.DeclareIdent("Title", filtering.TypeString),
	)
	if err != nil {
		panic(fmt.Sprintf("declare filter fields: %v", err))
	}
	return declarations
}

// listSongs queries Tracks with AIP-160 filtering, AIP-132 ordering, and pagination.
func listSongs(ctx context.Context, client *spanner.Client, req ListSongsRequest) (*ListSongsResponse, error) {
	declarations := songDeclarations()

	// Build SELECT columns.
	selectExpr := spansql.Select{
		List: []spansql.Expr{
			spansql.ID("SongId"),
			spansql.ID("Title"),
			spansql.ID("Artist"),
			spansql.ID("Genre"),
			spansql.ID("Year"),
		},
		From: []spansql.SelectFrom{
			spansql.SelectFromTable{Table: "Tracks"},
		},
	}

	params := map[string]any{}

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
		selectExpr.Where = whereExpr
		maps.Copy(params, filterParams)
	}

	// Parse and transpile order_by.
	var orderExprs []spansql.Order
	if req.OrderBy != "" {
		var ob ordering.OrderBy
		if err := ob.UnmarshalString(req.OrderBy); err != nil {
			return nil, fmt.Errorf("parse order_by: %w", err)
		}
		orderExprs = spanordering.TranspileOrderBy(ob)
	}
	// Always include SongId as tiebreaker for deterministic ordering.
	orderExprs = append(orderExprs, spansql.Order{Expr: spansql.ID("SongId")})

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

	// Build full query with LIMIT+1 for next page detection.
	query := spansql.Query{
		Select: selectExpr,
		Order:  orderExprs,
		Limit:  spansql.IntegerLiteral(pageSize + 1),
	}
	if offset > 0 {
		query.Offset = spansql.IntegerLiteral(offset)
	}

	stmt := spanner.Statement{
		SQL:    query.SQL(),
		Params: params,
	}

	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var songs []Song
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		var s Song
		if err := row.Columns(&s.SongId, &s.Title, &s.Artist, &s.Genre, &s.Year); err != nil {
			return nil, fmt.Errorf("scan columns: %w", err)
		}
		songs = append(songs, s)
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

// TestListFilter demonstrates AIP-132 List with AIP-160 filtering backed by Spanner.
func TestListFilter(t *testing.T) {
	ctx := context.Background()
	applySchema(t, ctx, "list_filter.sql")
	client := newClient(t, ctx)
	applySeed(t, ctx, client, "list_filter.sql")

	t.Run("no filter returns all songs", func(t *testing.T) {
		resp, err := listSongs(ctx, client, ListSongsRequest{})
		assert.NilError(t, err)
		assert.Equal(t, len(resp.Songs), 10)
		assert.Equal(t, resp.NextPageToken, "")
	})

	t.Run("filter by genre", func(t *testing.T) {
		resp, err := listSongs(ctx, client, ListSongsRequest{
			Filter: `Genre = "Rock"`,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(resp.Songs), 4)
		for _, s := range resp.Songs {
			assert.Equal(t, s.Genre, "Rock")
		}
	})

	t.Run("filter with AND", func(t *testing.T) {
		resp, err := listSongs(ctx, client, ListSongsRequest{
			Filter: `Genre = "Rock" AND Year > 1975`,
		})
		assert.NilError(t, err)
		// Smells Like Teen Spirit (1991) and Hotel California (1977).
		assert.Equal(t, len(resp.Songs), 2)
		for _, s := range resp.Songs {
			assert.Equal(t, s.Genre, "Rock")
			assert.Assert(t, s.Year > 1975, "expected year > 1975, got %d", s.Year)
		}
	})

	t.Run("filter with OR", func(t *testing.T) {
		resp, err := listSongs(ctx, client, ListSongsRequest{
			Filter: `Genre = "Rock" OR Genre = "Jazz"`,
		})
		assert.NilError(t, err)
		// 4 Rock + 3 Jazz = 7.
		assert.Equal(t, len(resp.Songs), 7)
		for _, s := range resp.Songs {
			assert.Assert(t, s.Genre == "Rock" || s.Genre == "Jazz",
				"expected Rock or Jazz, got %s", s.Genre)
		}
	})

	t.Run("order by year ascending", func(t *testing.T) {
		resp, err := listSongs(ctx, client, ListSongsRequest{
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
		resp, err := listSongs(ctx, client, ListSongsRequest{
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
			resp, err := listSongs(ctx, client, ListSongsRequest{
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
		// 10 songs with page_size=3 → 4 pages (3+3+3+1).
		assert.Equal(t, pages, 4)
		assert.Equal(t, len(allSongs), 10)
		// Verify deterministic ordering — SongIds should be 1..10.
		for i, s := range allSongs {
			assert.Equal(t, s.SongId, int64(i+1))
		}
	})

	t.Run("invalid filter returns error", func(t *testing.T) {
		_, err := listSongs(ctx, client, ListSongsRequest{
			Filter: `Genre = 123`,
		})
		assert.Assert(t, err != nil, "expected error for type mismatch filter")
	})
}
