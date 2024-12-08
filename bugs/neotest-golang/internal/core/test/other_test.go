//go:build integration

package test

import (
	"errors"
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	log.Fatal(errors.New("error in test main - bug"))
}
