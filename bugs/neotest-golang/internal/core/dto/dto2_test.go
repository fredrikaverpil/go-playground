//go:build integration

package dto_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fredrikaverpil/go-playground/bugs/neotest-golang/internal/core/dto"
	"github.com/stretchr/testify/require"
)

func TestSomething2(t *testing.T) {
	time.Sleep(10 * time.Second)
	p := &dto.Person{
		Name: "John",
		Age:  30,
	}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	p2 := new(dto.Person)
	require.NoError(t, json.Unmarshal(b, p2))
	require.Equal(t, p, p2)
}
