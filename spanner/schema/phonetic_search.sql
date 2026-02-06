CREATE TABLE Artists (
    ArtistId INT64 NOT NULL,
    FirstName STRING(1024),
    LastName STRING(1024),
    Genre STRING(256),
    FirstNameSoundex STRING(MAX) AS (LOWER(SOUNDEX(FirstName))),
    FirstNameSoundex_Tokens TOKENLIST AS (TOKEN(FirstNameSoundex)) HIDDEN
) PRIMARY KEY (ArtistId);

CREATE SEARCH INDEX ArtistsPhoneticIndex
    ON Artists(FirstNameSoundex_Tokens)
    STORING (FirstName, LastName, Genre, FirstNameSoundex);
