package httptransport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathnamePattern(t *testing.T) {
	tt := require.New(t)

	p := NewPathnamePattern("/users/:userID/repos/:repoID")

	tt.Equal(&PathnamePattern{
		parts: []string{"users", ":userID", "repos", ":repoID"},
		idxKeys: map[int]string{
			1: "userID",
			3: "repoID",
		},
	}, p)

	params, err := p.Parse("/users/1/repos/2")
	tt.NoError(err)

	tt.Equal("1", params.ByName("userID"))
	tt.Equal("2", params.ByName("repoID"))

	pathname := p.Stringify(params)
	tt.Equal("/users/1/repos/2", pathname)

	pathname2 := p.Stringify(ParamsFromMap(map[string]string{
		"userID": "1",
	}))
	tt.Equal("/users/1/repos/-", pathname2)

	{
		_, err := p.Parse("/not-match")
		tt.Error(err)
	}

	{
		_, err := p.Parse("/users/1/stars/1")
		tt.Error(err)
	}
}

func TestPathnamePatternWithoutParams(t *testing.T) {
	tt := require.New(t)

	p := NewPathnamePattern("/auth/user")

	tt.Equal(&PathnamePattern{
		parts:   []string{"auth", "user"},
		idxKeys: map[int]string{},
	}, p)

	{
		params, err := p.Parse("/auth/user")
		tt.NoError(err)
		tt.Len(params, 0)

		pathname := p.Stringify(params)
		tt.Equal("/auth/user", pathname)
	}
}
