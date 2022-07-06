### Multicall
**NOTE**: This repo is a fork of [https://github.com/Alethio/web3-multicall-go](https://github.com/Alethio/web3-multicall-go).
The main difference between this fork and the original is that this fork supports only the [original MakerDAO](https://github.com/makerdao/multicall)
contract ABI.

Wrapper for [Multicall](https://github.com/bowd/multicall) which batches calls to contract
view functions into a single call and reads all the state in one EVM round-trip.

### Usage

The library is used in conjunction with [web3-go](https://github.com/Alethio/web3-go), and the first parameter to `multicall.New` is an `ethrpc.ETHInterface` as defined in the package.

#### Initialization

This library requires the [MakerDAO multicall](https://github.com/makerdao/multicall) repo to be deployed on the target
chain. 

#### Calling

```go
package main

import (
    "fmt"
    "log"
    "math/big"
    "os"

    "github.com/TimelyToga/web3-multicall-go/multicall"
    "github.com/alethio/web3-go/ethrpc"
    "github.com/alethio/web3-go/ethrpc/provider/httprpc"
    "github.com/ethereum/go-ethereum/common"
)

// PolygonMulticallAddress was retrieved from official MakerDAO repo: https://github.com/makerdao/multicall
const PolygonMulticallAddress = "0x11ce4B23bD875D7F5C6a31084f55fDe1e9A87507"

const (
    balanceOfMethod = "balanceOf(address,uint256)(uint256)"
    ownerOfMethod   = "ownerOf(uint256)(address)"
)

// ComposeLDA_ID composes an LDA ID from it's composite parts
func ComposeLDA_ID(tierID, tokenID *big.Int) *big.Int {
    return new(big.Int).Add(new(big.Int).Lsh(tierID, 128), tokenID)
}

func main() {
    rpcURL := os.Getenv("RPC_URL")
    if rpcURL == "" {
        log.Fatal("RPC_URL must be set")
    }

    provider, err := httprpc.New(rpcURL)
    if err != nil {
        log.Fatal(err)
    }

    eth, err := ethrpc.New(provider)
    if err != nil {
        log.Fatal(err)
    }

    mc, err := multicall.New(eth, multicall.ContractAddress(PolygonMulticallAddress))
    fmt.Printf("%#v\n", mc)

    vcs := multicall.ViewCalls{}
    ids := []string{}

    tier := big.NewInt(3)
    for a := 0; a < 100; a++ {
        ldaID := ComposeLDA_ID(tier, big.NewInt(int64(a+1)))
        call := multicall.NewViewCall(
            fmt.Sprintf("owner-%d", a),
            "0x7c885c4bFd179fb59f1056FBea319D579A278075",
            ownerOfMethod,
            []interface{}{ldaID.String()},
        )
        ids = append(ids, ldaID.String())

        vcs = append(vcs, call)
    }

    // Default block parameter
    //block := "0x196166a"
    block := "latest"
    res, err := mc.Call(vcs, block)
    if err != nil {
        panic(err)
    }

    for a := 0; a < 100; a++ {
        key := fmt.Sprintf("owner-%d", a)
        value := res.Calls[key]
        ownerAddress := value.Decoded[0].(common.Address)
        ldaID := ids[a]
        fmt.Println(key, ":", ldaID, ownerAddress.String())
    }
}

```

This example shows a multicall query of the `ownerOf()` function in the Royal1155LDA contract for a 100 tokens at once.