CREATE TABLE Singers (
    SingerId INT64,
    FirstName STRING(1024),
    LastName STRING(1024),
    Metadata JSON
) PRIMARY KEY (SingerId);
