// Market buy (FOK). Amount is denominated in USDC. Requires resting asks on
// the book to fill against.
//
//	export PK=<hex private key>
//	export CHAIN_ID=137
//	export CLOB_API_URL=https://clob.polymarket.com
//	export CLOB_API_KEY=<key>
//	export CLOB_SECRET=<secret>
//	export CLOB_PASS_PHRASE=<passphrase>
//	export TOKEN_ID=<CTF token id>
//	go run ./examples/market_buy
//
// Swap OrderTypeFOK for OrderTypeFAK to fill partially instead of all-or-nothing.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jeffWu90/go-clob-client-v2/client"
	"github.com/jeffWu90/go-clob-client-v2/signing"
	"github.com/jeffWu90/go-clob-client-v2/types"
)

func main() {
	signer, err := signing.NewPrivateKeySigner(mustEnv("PK"))
	if err != nil {
		log.Fatalf("signer: %v", err)
	}

	chain := types.ChainAmoy
	if v := os.Getenv("CHAIN_ID"); v != "" {
		id, _ := strconv.Atoi(v)
		chain = types.Chain(id)
	}

	c, err := client.New(client.Options{
		Host:   envOr("CLOB_API_URL", "https://clob.polymarket.com"),
		Chain:  chain,
		Signer: signer,
		Creds: &types.ApiKeyCreds{
			Key:        mustEnv("CLOB_API_KEY"),
			Secret:     mustEnv("CLOB_SECRET"),
			Passphrase: mustEnv("CLOB_PASS_PHRASE"),
		},
	})
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	resp, err := c.CreateAndPostMarketOrder(
		context.Background(),
		types.UserMarketOrderV2{
			TokenID:   mustEnv("TOKEN_ID"),
			Amount:    100, // USDC
			Side:      types.SideBuy,
			OrderType: types.OrderTypeFOK,
		},
		types.CreateOrderOptions{TickSize: types.TickSize01},
		types.OrderTypeFOK,
		false, // deferExec
	)
	if err != nil {
		log.Fatalf("post market: %v", err)
	}
	fmt.Printf("%+v\n", resp)
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is required", k)
	}
	return v
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
