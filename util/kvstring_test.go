package util

import (
	"testing"
)

func TestMapSSToS(t *testing.T) {
	if a := MapSSToS(map[string]string{"a": "b"}); a != "{\"a\":\"b\"}" {
		t.Error(a)
	}
	if a := MapSSToS(map[string]string{"": ""}); a != "{}" {
		t.Error(a)
	}
}
