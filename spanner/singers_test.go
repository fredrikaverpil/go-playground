package main

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"gotest.tools/v3/assert"
)

type Metadata struct {
	Age  int    `json:"age"`
	City string `json:"city"`
}

type Artist struct {
	SingerID  int64
	FirstName string
	LastName  string
	Metadata  Metadata
}

func TestSingersInsert(t *testing.T) {
	ctx := context.Background()

	// Set up the Spanner client.
	client, err := spanner.NewClient(ctx, db, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	adminClient, err := database.NewDatabaseAdminClient(ctx, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to create database admin client: %v", err)
	}
	defer adminClient.Close()

	op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database: db,
		Statements: []string{
			`CREATE TABLE Singers (
          SingerId INT64,
          FirstName STRING(1024),
          LastName STRING(1024),
          Metadata JSON
       ) PRIMARY KEY (SingerId)

      `,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := op.Wait(ctx); err != nil {
		log.Fatal(err)
	}

	_, err = client.Apply(ctx, []*spanner.Mutation{
		spanner.Insert(
			"Singers",
			[]string{"SingerId", "FirstName", "LastName", "Metadata"},
			[]interface{}{1, "Marc", "Richards", `{"age": 30, "city": "New York"}`},
		),
		spanner.Insert("Singers", []string{"SingerId", "FirstName", "LastName"}, []interface{}{2, "Catalina", "Smith"}),
		spanner.Insert("Singers", []string{"SingerId", "FirstName", "LastName"}, []interface{}{3, "Alice", "Trentor"}),
		spanner.Insert("Singers", []string{"SingerId", "FirstName", "LastName"}, []interface{}{4, "Lea", "Martin"}),
		spanner.Insert("Singers", []string{"SingerId", "FirstName", "LastName"}, []interface{}{5, "David", "Lomond"}),
	})
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	expected := []Artist{
		{SingerID: 1, FirstName: "Marc", LastName: "Richards", Metadata: Metadata{Age: 30, City: "New York"}},
		{SingerID: 2, FirstName: "Catalina", LastName: "Smith", Metadata: Metadata{}},
		{SingerID: 3, FirstName: "Alice", LastName: "Trentor", Metadata: Metadata{}},
		{SingerID: 4, FirstName: "Lea", LastName: "Martin", Metadata: Metadata{}},
		{SingerID: 5, FirstName: "David", LastName: "Lomond", Metadata: Metadata{}},
	}
	got := make([]Artist, 0, len(expected))

	iter := client.Single().Query(ctx, spanner.NewStatement(`
    SELECT SingerId, FirstName, LastName, Metadata 
    FROM Singers 
    ORDER BY SingerId
  `))
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read data: %v", err)
		}
		var singerID int64
		var firstName, lastName string
		var spannerMetadata spanner.NullJSON
		if err := row.Columns(&singerID, &firstName, &lastName, &spannerMetadata); err != nil {
			log.Fatalf("Failed to read data: %v", err)
		}
		var metadata Metadata
		if spannerMetadata.Valid {
			jsonBytes, err := json.Marshal(spannerMetadata.Value)
			if err != nil {
				log.Fatalf("Failed to marshal JSON value: %v", err)
			}
			err = json.Unmarshal(jsonBytes, &metadata)
			if err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}
		}
		got = append(got, Artist{SingerID: singerID, FirstName: firstName, LastName: lastName, Metadata: metadata})
	}

	assert.DeepEqual(t, got, expected)
}
