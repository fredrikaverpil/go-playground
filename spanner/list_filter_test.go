package main

import (
	"fmt"

	"go.einride.tech/aip/filtering"
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
