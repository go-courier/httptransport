package transformers

import (
	"testing"
)

func TestTransformerCache(t *testing.T) {
	for name, tf := range TransformerMgrDefault.transformerSet {
		t.Log(name, tf)
	}
}
