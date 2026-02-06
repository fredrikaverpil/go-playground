CREATE TABLE Tracks (
    SongId INT64 NOT NULL,
    Title STRING(1024),
    Artist STRING(1024),
    Genre STRING(256),
    Year INT64
) PRIMARY KEY (SongId);
