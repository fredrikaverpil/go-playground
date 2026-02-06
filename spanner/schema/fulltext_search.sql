CREATE TABLE Songs (
    SongId INT64 NOT NULL,
    Title STRING(1024),
    Artist STRING(1024),
    Description STRING(MAX),
    Title_Tokens TOKENLIST AS (TOKENIZE_FULLTEXT(Title)) HIDDEN,
    Description_Tokens TOKENLIST AS (TOKENIZE_FULLTEXT(Description)) HIDDEN
) PRIMARY KEY (SongId);

CREATE SEARCH INDEX SongsFullTextIndex
    ON Songs(Title_Tokens, Description_Tokens)
    STORING (Title, Artist, Description);
