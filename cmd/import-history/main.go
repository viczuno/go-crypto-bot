package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/viczuno/go-crypto-bot/internal/platform"
	"log"
	"os"
	"time"
)

func main() {
	source := flag.String("sqlite", "archive/crypto_history.db", "SQLite source")
	history := flag.String("json", "archive/history", "legacy JSON directory")
	mode := flag.String("mode", "dry-run", "dry-run, import, verify")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	records, e := platform.ReadArchive(*source, *history)
	if e != nil {
		log.Fatal(e)
	}
	if *mode == "dry-run" {
		r, e := platform.Import(ctx, nil, records, false)
		if e != nil {
			log.Fatal(e)
		}
		json.NewEncoder(os.Stdout).Encode(r)
		return
	}
	if *mode != "import" && *mode != "verify" {
		log.Fatal("invalid mode")
	}
	c, e := platform.Connect(ctx)
	if e != nil {
		log.Fatal(e)
	}
	defer c.Close(ctx)
	r, e := platform.Import(ctx, c, records, *mode == "verify")
	if e != nil {
		log.Fatal(e)
	}
	json.NewEncoder(os.Stdout).Encode(r)
}
