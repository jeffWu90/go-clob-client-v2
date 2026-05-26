// Read-only example: fetch the orderbook for a token. No auth required.
//
//	export CLOB_API_URL=https://clob.polymarket.com
//	export TOKEN_ID=<CTF token id>
//	go run ./examples/get_orderbook
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jeffWu90/go-clob-client-v2/client"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

func main() {
	host := envOr("CLOB_API_URL", "https://clob.polymarket.com")
	tokenID := os.Getenv("TOKEN_ID")
	if tokenID == "" {
		log.Fatal("TOKEN_ID is required")
	}

	c, err := client.New(client.Options{Host: host, Chain: types.ChainPolygon})
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	book, err := c.GetOrderBook(context.Background(), tokenID)
	if err != nil {
		log.Fatalf("get orderbook: %v", err)
	}

	fmt.Printf("market=%s asset=%s tick=%s neg_risk=%v\n", book.Market, book.AssetID, book.TickSize, book.NegRisk)
	fmt.Println("bids (highest first):")
	for _, b := range book.Bids {
		fmt.Printf("  %s @ %s\n", b.Size, b.Price)
	}
	fmt.Println("asks (lowest first):")
	for _, a := range book.Asks {
		fmt.Printf("  %s @ %s\n", a.Size, a.Price)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
