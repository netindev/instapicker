package main

import (
	"fmt"
	"log"

	"github.com/netindev/instapicker/internal/browser"
	"github.com/netindev/instapicker/internal/config"
	"github.com/netindev/instapicker/internal/instagram"
)

func main() {
	cfg := config.Load()

	b, err := browser.Start(cfg.Headless)
	if err != nil {
		log.Fatalf("failed to start browser: %v", err)
	}
	defer b.Close()

	client := instagram.NewClient(b)

	if err := client.Login(cfg.Username, cfg.Password); err != nil {
		log.Fatalf("login failed: %v", err)
	}

	comments, err := client.GetComments(cfg.URL)
	if err != nil {
		log.Fatalf("failed to fetch comments: %v", err)
	}

	if err := instagram.WriteResult(comments); err != nil {
		log.Fatalf("failed to write result: %v", err)
	}

	fmt.Println(comments)
}
