// Bootstrap example: mint or derive L2 (HMAC) API credentials with a wallet.
//
//	export PK=<hex private key>
//	export CHAIN_ID=137                    # 137 = Polygon, 80002 = Amoy
//	export CLOB_API_URL=https://clob.polymarket.com
//	go run ./examples/create_api_key
//
// The returned credentials are what every other authenticated endpoint needs;
// store them somewhere safe (they CANNOT be recovered once lost).
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
	pk := os.Getenv("PK")
	if pk == "" {
		log.Fatal("PK is required (hex private key)")
	}
	signer, err := signing.NewPrivateKeySigner(pk)
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
	})
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	creds, err := c.CreateOrDeriveApiKey(context.Background(), 0)
	if err != nil {
		log.Fatalf("create/derive: %v", err)
	}

	fmt.Println("Save these credentials securely. They CANNOT be recovered.")
	fmt.Println("CLOB_API_KEY=" + creds.Key)
	fmt.Println("CLOB_SECRET=" + creds.Secret)
	fmt.Println("CLOB_PASS_PHRASE=" + creds.Passphrase)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
