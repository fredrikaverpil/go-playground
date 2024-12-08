package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
	p := &Person{
		Name: "John",
		Age:  30,
	}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	p2 := new(Person)
	require.NoError(t, json.Unmarshal(b, p2))
	require.Equal(t, p, p2)
}
