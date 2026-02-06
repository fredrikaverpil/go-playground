CREATE TABLE AlbumsNgramMin1 (
    AlbumId INT64 NOT NULL,
    Title STRING(1024),
    Title_Tokens TOKENLIST AS (
        TOKENIZE_SUBSTRING(Title, ngram_size_min=>1, ngram_size_max=>3,
            relative_search_types=>["word_prefix", "word_suffix"])
    ) HIDDEN
) PRIMARY KEY (AlbumId);

CREATE SEARCH INDEX AlbumsNgramMin1Index
    ON AlbumsNgramMin1(Title_Tokens)
    STORING (Title);

CREATE TABLE AlbumsNgramMin2 (
    AlbumId INT64 NOT NULL,
    Title STRING(1024),
    Title_Tokens TOKENLIST AS (
        TOKENIZE_SUBSTRING(Title, ngram_size_min=>2, ngram_size_max=>3,
            relative_search_types=>["word_prefix", "word_suffix"])
    ) HIDDEN
) PRIMARY KEY (AlbumId);

CREATE SEARCH INDEX AlbumsNgramMin2Index
    ON AlbumsNgramMin2(Title_Tokens)
    STORING (Title);
