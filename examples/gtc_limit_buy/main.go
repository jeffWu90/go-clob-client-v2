// Resting limit buy. The order sits in the book at the requested price until
// filled or cancelled.
//
//	export PK=<hex private key>
//	export CHAIN_ID=137
//	export CLOB_API_URL=https://clob.polymarket.com
//	export CLOB_API_KEY=<key>
//	export CLOB_SECRET=<secret>
//	export CLOB_PASS_PHRASE=<passphrase>
//	export TOKEN_ID=<CTF token id>
//	go run ./examples/gtc_limit_buy
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

	resp, err := c.CreateAndPostOrder(
		context.Background(),
		types.UserOrderV2{
			TokenID: mustEnv("TOKEN_ID"),
			Price:   0.4,
			Size:    100,
			Side:    types.SideBuy,
		},
		types.CreateOrderOptions{TickSize: types.TickSize01},
		types.OrderTypeGTC,
		false, // postOnly
		false, // deferExec
	)
	if err != nil {
		log.Fatalf("post: %v", err)
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
