package git

import (
	"testing"
)

func TestChangedFileSetMatches(t *testing.T) {
	t.Run("nil set matches everything", func(t *testing.T) {
		var nilSet ChangedFileSet
		if !nilSet.Matches("/any/path") {
			t.Error("nil set should match everything")
		}
	})

	t.Run("empty set matches nothing", func(t *testing.T) {
		s := make(ChangedFileSet)
		if s.Matches("/any/path") {
			t.Error("empty set should not match")
		}
	})

	t.Run("set with specific path", func(t *testing.T) {
		s := ChangedFileSet{"/tmp/test.go": true}
		if !s.Matches("/tmp/test.go") {
			t.Error("should match existing path")
		}
		if s.Matches("/tmp/other.go") {
			t.Error("should not match missing path")
		}
	})
}
