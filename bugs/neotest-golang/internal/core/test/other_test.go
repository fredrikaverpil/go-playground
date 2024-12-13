//go:build integration

package test

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestSomething(t *testing.T) {
	require.Equal(t, 1, 1)
}

func Test(t *testing.T) {
	b := make([]byte, 1000)
	_, err := rand.Read(b)
	require.NoError(t, err)
	os.Stdout.Write(b)
}
