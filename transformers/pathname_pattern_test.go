package transformers

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestPathnamePattern(t *testing.T) {
	p := NewPathnamePattern("/users/:userID/repos/:repoID")

	NewWithT(t).Expect(p).To(Equal(&PathnamePattern{
		parts: []string{"users", ":userID", "repos", ":repoID"},
		idxKeys: map[int]string{
			1: "userID",
			3: "repoID",
		},
	}))

	t.Run("parse success", func(t *testing.T) {
		params, err := p.Parse("/users/1/repos/2")

		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(params.ByName("userID")).To(Equal("1"))
		NewWithT(t).Expect(params.ByName("repoID")).To(Equal("2"))

		NewWithT(t).Expect(p.Stringify(params)).To(Equal("/users/1/repos/2"))
	})

	t.Run("stringify with empty missing params", func(t *testing.T) {
		NewWithT(t).Expect(p.Stringify(ParamsFromMap(map[string]string{
			"userID": "1",
		}))).To(Equal("/users/1/repos/-"))
	})

	t.Run("parse failed for path which not matched", func(t *testing.T) {
		_, err := p.Parse("/not-match")
		NewWithT(t).Expect(err).NotTo(BeNil())
	})

	t.Run("parse failed for path which not full matched", func(t *testing.T) {
		_, err := p.Parse("/users/1/stars/1")
		NewWithT(t).Expect(err).NotTo(BeNil())
	})
}

func TestPathnamePatternWithoutParams(t *testing.T) {
	p := NewPathnamePattern("/auth/user")

	NewWithT(t).Expect(p).To(Equal(&PathnamePattern{
		parts:   []string{"auth", "user"},
		idxKeys: map[int]string{},
	}))

	{
		params, err := p.Parse("/auth/user")
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(params).To(HaveLen(0))

		NewWithT(t).Expect(p.Stringify(params)).To(Equal("/auth/user"))
	}
}
