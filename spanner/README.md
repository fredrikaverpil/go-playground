# Spanner playground

A playground for learning Spanner using the emulator.

## Quickstart

```bash
go test -v -count=1 -race ./...
```

The `-count=1` avoids caching.

### Benchmarks

```bash
go test -bench=. -benchmem -count=1 -run='^$' ./...
```

Each benchmark compares the native `spanner.Client` against `database/sql` (via
[go-sql-spanner](https://github.com/googleapis/go-sql-spanner)).

Results on Apple M2 (emulator, `count=1`):

| Benchmark      | Query                            | `spanner` ns/op | `database/sql` ns/op | Ratio |
| -------------- | -------------------------------- | --------------: | -------------------: | ----: |
| Singers        | `SELECT` with JSON column        |       2,576,532 |            3,994,969 | 1.55x |
| FullTextSearch | `SEARCH()` on full-text tokens   |       2,645,714 |            3,981,792 | 1.51x |
| FuzzySearch    | `SEARCH_NGRAMS` + `SCORE_NGRAMS` |       2,676,155 |            4,134,266 | 1.54x |
| PhoneticSearch | `SOUNDEX`-based equality filter  |       2,576,991 |            3,971,287 | 1.54x |
| ListFilter     | Parameterized `WHERE` clause     |       2,529,799 |            3,891,600 | 1.54x |

The native client is consistently ~1.5x faster than `database/sql`, which adds
overhead from the generic `sql.DB` abstraction layer.

## How it works

`main_test.go` contains `TestMain`, which:

1. Starts a Spanner emulator Docker container
2. Waits for the emulator ports to be ready
3. Creates the Spanner instance and database
4. Runs all tests
5. Stops the emulator container

Each test uses shared helpers from `testhelpers_test.go` to apply schema (DDL)
and seed data (DML) from embedded SQL files before running queries. Schema and
seed files are named after the experiment (1:1 mapping):

| Experiment       | Native (`_spanner`)               | database/sql (`_sql`)         | Benchmark (`_bench`)            | Schema                       | Seed                       |
| ---------------- | --------------------------------- | ----------------------------- | ------------------------------- | ---------------------------- | -------------------------- |
| Singers          | `singers_spanner_test.go`         | `singers_sql_test.go`         | `singers_bench_test.go`         | `schema/singers.sql`         | `seed/singers.sql`         |
| Full-text search | `fulltext_search_spanner_test.go` | `fulltext_search_sql_test.go` | `fulltext_search_bench_test.go` | `schema/fulltext_search.sql` | `seed/fulltext_search.sql` |
| Fuzzy search     | `fuzzy_search_spanner_test.go`    | `fuzzy_search_sql_test.go`    | `fuzzy_search_bench_test.go`    | `schema/fuzzy_search.sql`    | `seed/fuzzy_search.sql`    |
| Phonetic search  | `phonetic_search_spanner_test.go` | `phonetic_search_sql_test.go` | `phonetic_search_bench_test.go` | `schema/phonetic_search.sql` | `seed/phonetic_search.sql` |
| List filter      | `list_filter_spanner_test.go`     | `list_filter_sql_test.go`     | `list_filter_bench_test.go`     | `schema/list_filter.sql`     | `seed/list_filter.sql`     |
| N-gram bench     | —                                 | —                             | `ngram_bench_test.go`           | `schema/ngram_bench.sql`     | `seed/ngram_bench.sql`     |

Experiments with shared types also have an unsuffixed `_test.go` file (e.g.
`singers_test.go`, `list_filter_test.go`) containing only type definitions and
declarations.

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

#### `TOKENIZE_SUBSTRING` configuration

Google recommends `ngram_size_min=>2, ngram_size_max=>3` as a starting point for
fuzzy search / typo matching. Key guidance:

- **Avoid `ngram_size_min=1`** — single-character n-grams match too many
  documents and bloat the index. From the docs: "We don't recommend one character
  n-grams because they could match a very large number of documents."
- **Substring indexes use 10-30x more storage** than full-text indexes over the
  same data. The overhead grows as the gap between `ngram_size_min` and
  `ngram_size_max` widens.
- **`ngram_size_min=>4, ngram_size_max=>6`** is recommended for substring search
  (exact substring matching, not fuzzy).
- **`short_tokens_only_for_anchors=>TRUE`** reduces token count when only
  prefix/suffix matching is needed (requires `relative_search_types` to be set
  to a prefix or suffix mode).
- Only increase `ngram_size_min` when you control the queries and can guarantee
  the minimum query length meets or exceeds `ngram_size_min`.

#### `SCORE_NGRAMS`

Uses Jaccard similarity over trigrams:
`shared_ngrams / (source_ngrams + query_ngrams - shared_ngrams)`. The
`algorithm` parameter only supports `"trigrams"` (default, and the only option).

- **Always use with `SEARCH_NGRAMS`** in a filter + rank pattern. Use the same
  query parameter in both functions.
- **Needs the source column** (not just the index), so include the source column
  in the search index's `STORING` clause to avoid a join with the base table.
- **Use an inner `LIMIT`** to avoid expensive queries when popular n-gram
  combinations are encountered.

Recommended query pattern:

```sql
SELECT AlbumId FROM (
  SELECT AlbumId, SCORE_NGRAMS(Title_Tokens, @p) AS score
  FROM Albums
  WHERE SEARCH_NGRAMS(Title_Tokens, @p)
  LIMIT 10000
)
ORDER BY score DESC
LIMIT 10
```

#### Spanner search approaches

| Approach         | Function            | Tolerates typos? | Partial words? | Use case                                        |
| ---------------- | ------------------- | ---------------- | -------------- | ----------------------------------------------- |
| Fuzzy search     | `SEARCH_NGRAMS`     | Yes              | Yes            | User-facing search where input has typos        |
| Substring search | `SEARCH_SUBSTRING`  | No               | Yes            | Exact substring filter (indexed `LIKE '%foo%'`) |
| Full-text search | `SEARCH`            | No               | No             | Keyword/phrase search over documents            |
| Phonetic search  | `SOUNDEX` + `TOKEN` | Sound-alike only | No             | Name matching (Steven/Stephen)                  |

For user-facing search boxes with free-form text input, `SEARCH_NGRAMS` is the
best fit due to its typo tolerance. `SEARCH_SUBSTRING` would be the alternative
when input is always exact (e.g. copy-pasted identifiers, autocomplete
selections) — cheaper on index storage, more precise, but zero typo tolerance.
The n-gram size benchmark below focuses on `SEARCH_NGRAMS` for this reason.

Note: `relative_search_types` in `TOKENIZE_SUBSTRING` generates anchor tokens
for `SEARCH_SUBSTRING` positional matching (word prefix, suffix, phrase
adjacency). `SEARCH_NGRAMS` ignores these anchors, so enabling
`relative_search_types` when only using `SEARCH_NGRAMS` bloats the index
without benefit.

#### Full-text search vs fuzzy search

|                 | Full-text (`SEARCH` + `SCORE`) | Fuzzy (`SEARCH_NGRAMS` + `SCORE_NGRAMS`) |
| --------------- | ------------------------------ | ---------------------------------------- |
| Tokenization    | Words (`TOKENIZE_FULLTEXT`)    | Character n-grams (`TOKENIZE_SUBSTRING`) |
| Matching        | Exact words (with stemming)    | Approximate (shared n-grams)             |
| Typo tolerance  | No                             | Yes                                      |
| Partial words   | No                             | Yes                                      |
| Boolean queries | Yes (`OR`, `AND`)              | No                                       |
| Phrase matching | Yes                            | No                                       |
| Scoring         | TF-IDF based relevance         | Jaccard similarity over trigrams         |
| Index overhead  | Baseline                       | 10-30x more storage                      |

- [Find approximate matches with fuzzy search](https://docs.cloud.google.com/spanner/docs/full-text-search/fuzzy-search)
- [Perform a substring search](https://cloud.google.com/spanner/docs/full-text-search/substring-search)
- [Tokenization](https://docs.cloud.google.com/spanner/docs/full-text-search/tokenization)
- [Search functions reference](https://docs.cloud.google.com/spanner/docs/reference/standard-sql/search_functions)

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

### N-gram size benchmark

Compares `ngram_size_min=>1` vs `ngram_size_min=>2` vs `ngram_size_min=>3` in
`TOKENIZE_SUBSTRING` to measure the impact on `SEARCH_NGRAMS` + `SCORE_NGRAMS`
query performance. Uses three identical tables with different tokenization
configs and the same 50-row dataset. Only benchmarks the native `spanner.Client`
since the comparison is between tokenization parameters, not drivers.

`BenchmarkNgram` runs configs in ascending order (1, 2, 3) and
`BenchmarkNgramReversed` runs them in descending order (3, 2, 1). Running all
configs in a single process produces an ordering bias (see below), so for a fair
comparison run each config in isolation:

```bash
go test -bench='^BenchmarkNgram$/ngram_min_1' -benchmem -count=3 -run='^$' ./...
go test -bench='^BenchmarkNgram$/ngram_min_2' -benchmem -count=3 -run='^$' ./...
go test -bench='^BenchmarkNgram$/ngram_min_3' -benchmem -count=3 -run='^$' ./...
```

**Emulator caveat:** The Spanner emulator (Java-based) degrades under sustained
load — ns/op increases monotonically across consecutive benchmark iterations
within a single process. When all configs run together, whichever sub-benchmark
runs first always appears faster. Running each config in its own process (fresh
emulator) eliminates this bias.

#### Isolated results (fair comparison)

Each config in its own process with a fresh emulator. Results on Apple M2
(emulator, `count=3`):

| Config        | Run 1 ns/op | Run 2 ns/op | Run 3 ns/op |    B/op | allocs/op |
| ------------- | ----------: | ----------: | ----------: | ------: | --------: |
| `ngram_min_1` |   3,029,308 |   3,848,926 |   4,591,408 | ~21,460 |       328 |
| `ngram_min_2` |   2,960,212 |   3,869,940 |   4,803,368 | ~21,460 |       328 |
| `ngram_min_3` |   2,921,224 |   3,831,745 |   4,683,471 | ~21,460 |       328 |

**No measurable difference** between `ngram_size_min=1`, `=2`, and `=3` on the
emulator. All three configs produce the same times within noise (~3.0ms →
~4.7ms) and identical memory allocations.

Note: within-process degradation (3ms → 4.7ms across `count=3`) still exists
but affects each config equally. A real-world difference, if any, would need to
be validated on Cloud Spanner with a larger dataset.

#### Combined results (showing ordering bias)

All configs in one process. Included to demonstrate why isolated runs are
necessary:

`BenchmarkNgram` (ascending: 1 → 2 → 3):

| Config                      | Run 1 ns/op | Run 3 ns/op |
| --------------------------- | ----------: | ----------: |
| `ngram_min_1` (runs first)  |   3,042,706 |   4,671,490 |
| `ngram_min_2` (runs second) |   5,211,144 |   6,074,381 |
| `ngram_min_3` (runs third)  |   6,525,386 |   7,143,420 |

`BenchmarkNgramReversed` (descending: 3 → 2 → 1):

| Config                      | Run 1 ns/op | Run 3 ns/op |
| --------------------------- | ----------: | ----------: |
| `ngram_min_3` (runs first)  |   7,635,329 |   8,202,486 |
| `ngram_min_2` (runs second) |   8,513,336 |   8,934,733 |
| `ngram_min_1` (runs third)  |   9,020,893 |   9,531,930 |

Whichever config runs first always gets the fastest times. `ngram_min_1` went
from 3.0ms (first) to 9.0ms (last) by simply changing the execution order.

- [Find approximate matches with fuzzy search](https://docs.cloud.google.com/spanner/docs/full-text-search/fuzzy-search)
- [Tokenization](https://docs.cloud.google.com/spanner/docs/full-text-search/tokenization)
