// Package royal is for utilities that should be shared by all golang apps at royal
package royal

import (
    "fmt"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/ethclient"
)

// EthClient is an interface that supports all the methods of backends.SimulatedBackend and ethclient.Client so mocking
// is way easier. This is interface that should be used in our code
type EthClient interface {
    ethereum.ChainStateReader
    ethereum.PendingStateReader
    ethereum.GasPricer
    ethereum.GasEstimator
    ethereum.TransactionReader
    ethereum.TransactionSender
    bind.ContractBackend

    // GetChainID is a helper method for retrieving the ChainId of the current chain
    GetChainID() uint64
}

// ethClient is an implementation of the EthClient interface. This interface is used so we have a layer of
// indirection between the ethclient.Client struct so we can more effectively mock our code and also simplify the
// places that we interact with an eth-compatible network
type ethClient struct {
    *ethclient.Client
    chainID uint64
}

// NewClient initializes a new connection to the given chainID given the rpcURL
func NewClient(chainID uint64, rpcURL string) (EthClient, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, fmt.Errorf("trouble creating ethClient for (chainID:%d): %s\n", chainID, err)
    }

    return &ethClient{
        Client:  client,
        chainID: chainID,
    }, nil
}

// GetChainID returns the chainID for this client
func (rec *ethClient) GetChainID() uint64 {
    return rec.chainID
}
