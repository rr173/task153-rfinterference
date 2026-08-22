package main

import (
	"context"
	"flag"
	"fmt"
	"task153-rfinterference/internal/demo"
	"task153-rfinterference/internal/httpapi"
	"task153-rfinterference/internal/service"
	"task153-rfinterference/internal/store"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var dbPath, addr string
	var smoke bool
	flag.StringVar(&dbPath, "db", "rfinterference.db", "SQLite database path")
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.BoolVar(&smoke, "smoke-test", false, "run demo self-check and exit")
	flag.Parse()
	if smoke {
		runSmoke()
		return
	}
	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	svc := service.New(db)
	if err := svc.Recover(context.Background()); err != nil {
		log.Fatal(err)
	}
	server := httpapi.Server(addr, httpapi.New(svc).Handler())
	log.Printf("rf interference attribution API listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
func runSmoke() {
	dir, err := os.MkdirTemp("", "rf-smoke-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	db, err := store.Open(filepath.Join(dir, "smoke.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	svc := service.New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := demo.Import(ctx, svc); err != nil {
		log.Fatal(err)
	}
	report, err := demo.SelfCheck(ctx, svc)
	if err != nil || !report.OK {
		log.Fatalf("self-check failed: %+v %v", report, err)
	}
	fmt.Printf("smoke-test passed: %d event(s), %d fragment(s)\n", report.EventCount, report.FragmentCount)
}
