package demo

import (
	"context"
	"github.com/rr173/task153-rfinterference/internal/service"
	"github.com/rr173/task153-rfinterference/internal/store"
	"path/filepath"
	"testing"
)

func TestImportAndSelfCheck(t *testing.T) {
	db, _ := store.Open(filepath.Join(t.TempDir(), "demo.db"))
	defer db.Close()
	svc := service.New(db)
	out, err := Import(context.Background(), svc)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 {
		t.Fatalf("got %d", len(out))
	}
	r, err := SelfCheck(context.Background(), svc)
	if err != nil || !r.OK {
		t.Fatalf("%+v %v", r, err)
	}
}
