# Spanner playground

A playground for learning Spanner using the emulator.

## Quickstart

```bash
go test -v -count=1 -race ./...
```

The `-count=1` avoids caching.

## How it works

`main_test.go` contains `TestMain`, which:

1. Starts a Spanner emulator Docker container
2. Waits for the emulator ports to be ready
3. Creates the Spanner instance and database
4. Runs all tests
5. Stops the emulator container

Each test uses shared helpers from `testhelpers_test.go` to apply schema (DDL)
and seed data (DML) from embedded SQL files before running queries. Schema and
seed files are named after the test that uses them (1:1 mapping):

| Test | Schema | Seed |
|---|---|---|
| `singers_test.go` | `schema/singers.sql` | `seed/singers.sql` |
| `fulltext_search_test.go` | `schema/fulltext_search.sql` | `seed/fulltext_search.sql` |
| `fuzzy_search_test.go` | `schema/fuzzy_search.sql` | `seed/fuzzy_search.sql` |
| `phonetic_search_test.go` | `schema/phonetic_search.sql` | `seed/phonetic_search.sql` |
| `list_filter_test.go` | `schema/list_filter.sql` | `seed/list_filter.sql` |

## Experiments

### Singers CRUD

Basic insert and query with JSON column support. Demonstrates `spanner.NullJSON`
handling for nullable JSON fields.

- [Work with JSON data](https://docs.cloud.google.com/spanner/docs/working-with-json)
- [JSON functions in GoogleSQL](https://docs.cloud.google.com/spanner/docs/reference/standard-sql/json_functions)

### Full-text search

Uses `TOKENIZE_FULLTEXT` to break text into searchable tokens, `SEARCH()` to
filter by token match, and `SCORE()` to rank results by relevance. Covers
single-word search, multi-column search, boolean OR queries, and scoring.

- [Full-text search overview](https://docs.cloud.google.com/spanner/docs/full-text-search)
- [Tokenization](https://docs.cloud.google.com/spanner/docs/full-text-search/tokenization)
- [Search functions in GoogleSQL](https://docs.cloud.google.com/spanner/docs/reference/standard-sql/search_functions)

### Fuzzy search (n-gram)

Uses `TOKENIZE_SUBSTRING` with configurable n-gram sizes to enable approximate
matching. `SEARCH_NGRAMS()` finds candidates sharing n-grams with the query, and
`SCORE_NGRAMS()` ranks by Jaccard similarity. Handles misspellings and partial
word matches.

- [Find approximate matches with fuzzy search](https://docs.cloud.google.com/spanner/docs/full-text-search/fuzzy-search)
- [Tokenization](https://docs.cloud.google.com/spanner/docs/full-text-search/tokenization)

### Phonetic search (SOUNDEX)

Uses `SOUNDEX()` to generate phonetic codes for names, enabling searches that
match different spellings of similar-sounding words (e.g. Steven/Stephen/Stefan,
Carl/Karl). The soundex code is tokenized with `TOKEN()` for use in a search
index.

- [Find approximate matches with fuzzy search](https://docs.cloud.google.com/spanner/docs/full-text-search/fuzzy-search)
- [String functions in GoogleSQL (`SOUNDEX`)](https://docs.cloud.google.com/spanner/docs/reference/standard-sql/string_functions)

### AIP-132 List with AIP-160 filtering

Parses [AIP-160](https://google.aip.dev/160) filter strings and transpiles them
to Spanner SQL `WHERE` clauses using
[`go.einride.tech/aip`](https://pkg.go.dev/go.einride.tech/aip) and
[`go.einride.tech/spanner-aip`](https://pkg.go.dev/go.einride.tech/spanner-aip).
Supports [AIP-132](https://google.aip.dev/132) ordering and offset-based
pagination with the limit+1 pattern for `next_page_token`.

- [AIP-132: Standard methods: List](https://google.aip.dev/132)
- [AIP-160: Filtering](https://google.aip.dev/160)
