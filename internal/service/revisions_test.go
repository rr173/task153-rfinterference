package service

import (
	"strings"
	"testing"

	"github.com/rr173/task153-rfinterference/internal/model"
)

func TestBuildRevisionNoteMarksArchiveAsFrozen(t *testing.T) {
	note := BuildRevisionNote(model.Event{Revision: 3, Frozen: true}, nil, nil, nil)
	if !note.Frozen || !strings.Contains(note.Summary, "frozen") {
		t.Fatalf("unexpected note: %+v", note)
	}
}
