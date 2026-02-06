CREATE TABLE Albums (
    AlbumId INT64 NOT NULL,
    Title STRING(1024),
    Artist STRING(1024),
    Title_Tokens TOKENLIST AS (
        TOKENIZE_SUBSTRING(Title, ngram_size_min=>2, ngram_size_max=>3,
            relative_search_types=>["word_prefix", "word_suffix"])
    ) HIDDEN
) PRIMARY KEY (AlbumId);

CREATE SEARCH INDEX AlbumsNgramIndex
    ON Albums(Title_Tokens)
    STORING (Title, Artist);
