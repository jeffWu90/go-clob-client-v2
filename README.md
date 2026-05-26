# go-clob-client-v2

Go client for the Polymarket CLOB V2. Functional parity with the TypeScript [`@polymarket/clob-client-v2`](https://github.com/Polymarket/clob-client-v2): L1 (EIP-712) and L2 (HMAC) auth, V2 order builder, full HTTP API surface, signature byte-identical to the upstream ethers / viem implementations.

## Install

```bash
go get github.com/jeffWu90/go-clob-client-v2
```

Requires Go 1.21+.

## Quick start

### Limit buy (GTC)

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/jeffWu90/go-clob-client-v2/client"
    "github.com/jeffWu90/go-clob-client-v2/signing"
    "github.com/jeffWu90/go-clob-client-v2/types"
)

func main() {
    signer, err := signing.NewPrivateKeySigner(os.Getenv("PK"))
    if err != nil {
        panic(err)
    }

    // Step 1: mint or fetch the L2 (HMAC) credentials with the wallet (L1 auth).
    bootstrap, _ := client.New(client.Options{
        Host:   "https://clob.polymarket.com",
        Chain:  types.ChainPolygon,
        Signer: signer,
    })
    creds, err := bootstrap.CreateOrDeriveApiKey(context.Background(), 0)
    if err != nil {
        panic(err)
    }

    // Step 2: build a fully-authenticated client with the credentials in hand.
    c, _ := client.New(client.Options{
        Host:   "https://clob.polymarket.com",
        Chain:  types.ChainPolygon,
        Signer: signer,
        Creds:  creds,
    })

    resp, err := c.CreateAndPostOrder(
        context.Background(),
        types.UserOrderV2{
            TokenID: "<CTF token id>",
            Price:   0.4,
            Side:    types.SideBuy,
            Size:    100,
        },
        types.CreateOrderOptions{TickSize: types.TickSize01},
        types.OrderTypeGTC,
        false, // postOnly
        false, // deferExec
    )
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", resp)
}
```

### Market buy (FOK)

```go
resp, err := c.CreateAndPostMarketOrder(
    ctx,
    types.UserMarketOrderV2{
        TokenID:   tokenID,
        Amount:    100, // USDC for a BUY; shares for a SELL
        Side:      types.SideBuy,
        OrderType: types.OrderTypeFOK,
    },
    types.CreateOrderOptions{TickSize: types.TickSize01},
    types.OrderTypeFOK,
    false, // deferExec
)
```

For FAK semantics swap `OrderTypeFOK` for `OrderTypeFAK` in both spots.

### Reading market data (no auth)

```go
c, _ := client.New(client.Options{
    Host:  "https://clob.polymarket.com",
    Chain: types.ChainPolygon,
})

book, _ := c.GetOrderBook(ctx, tokenID)
fmt.Println(book.Bids, book.Asks)

ts, _ := c.GetServerTime(ctx)
```

## Authentication

Two layers, both required for trading endpoints:

| Layer | Mechanism | Used for |
|---|---|---|
| **L1** | Wallet EIP-712 signature | minting / deriving API keys |
| **L2** | HMAC-SHA256 with API secret | every other authenticated endpoint |

L1 sign-in is a one-time setup. After `CreateOrDeriveApiKey` you keep the returned `*types.ApiKeyCreds` and pass it to subsequent `client.New` calls.

```go
creds := &types.ApiKeyCreds{
    Key:        os.Getenv("CLOB_API_KEY"),
    Secret:     os.Getenv("CLOB_SECRET"),
    Passphrase: os.Getenv("CLOB_PASS_PHRASE"),
}
c, _ := client.New(client.Options{
    Host: host, Chain: chain, Signer: signer, Creds: creds,
})
```

## Error handling

All API methods return `(result, error)`. On a non-2xx HTTP response the error is `*clob.ApiError`:

```go
import (
    "errors"

    clob "github.com/jeffWu90/go-clob-client-v2"
)

book, err := c.GetOrderBook(ctx, tokenID)
if err != nil {
    var apiErr *clob.ApiError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.Message) // e.g. "No orderbook exists for the requested token id"
        fmt.Println(apiErr.Status)  // 404
        fmt.Println(apiErr.Data)    // full decoded body
    }
}
```

Auth gating returns sentinel errors:

```go
var (
    clob.ErrL1AuthUnavailable // signer missing
    clob.ErrL2AuthUnavailable // creds missing
)
```

## Signers

```go
// In-process secp256k1 (EOA)
signer, _ := signing.NewPrivateKeySigner(privateKeyHex)

// Custom (HSM, remote signer, etc.) — implement the two-method Signer interface:
type Signer interface {
    Address() string
    SignTypedData(td *apitypes.TypedData) (string, error)
}
```

`PrivateKeySigner` accepts the hex key with or without the `0x` prefix.

## Package layout

| Package | Purpose |
|---|---|
| `client` | `ClobClient` — top-level API surface (~50 endpoints) |
| `signing` | EIP-712 + HMAC signers, `PrivateKeySigner` |
| `orderbuilder` | Order construction (limit + market), amount math, market price calc, wire encoding |
| `orderutils` | EIP-712 typed data + hashing for the CTF Exchange V2 |
| `headers` | `POLY_*` request-header builders |
| `httphelpers` | HTTP client with retry, error mapping |
| `types` | All response / request DTOs and enums (Chain, Side, OrderType, TickSize, …) |
| root `clob` | Endpoint constants, contract config, rounding helpers, `ApiError` |

## Differences from the TS client

- **V2 only**. V1 order builder, V1 signature type, and version-resolution helpers are not implemented. Use the TS client if you need V1 support.
- **Errors are returned, not thrown.** There is no `throwOnError` flag — every method returns `(result, error)`, and API failures surface as `*clob.ApiError`.
- **`OrderBuilder.getSigner()` dynamic resolver is omitted.** Implement the `signing.Signer` interface yourself if you need rotation.
- **Tests vs source.** HMAC and EIP-712 outputs are byte-identical to the TS test vectors. Order amount / market-price calculations reproduce the TS property-test invariants across the same tick / price grid.

## Examples

See [`examples/`](examples/) for runnable programs. Each one reads `PK`, `CHAIN_ID`, `CLOB_API_URL`, `CLOB_API_KEY`, `CLOB_SECRET`, `CLOB_PASS_PHRASE` from the environment.

```bash
go run ./examples/gtc_limit_buy
go run ./examples/market_buy
go run ./examples/get_orderbook
go run ./examples/create_api_key
```

## License

MIT.
