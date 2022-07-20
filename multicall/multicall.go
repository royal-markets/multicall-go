package multicall

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/common/hexutil"

    "github.com/TimelyToga/web3-multicall-go/royal"
)

type Multicall interface {
    CallRaw(calls ViewCalls, block string) (*Result, error)
    Call(calls ViewCalls, block string) (*Result, error)
    Contract() string
}

type multicall struct {
    client royal.EthClient
    config *Config
}

const (
    AggregateMethod = "0x252dba42"
    DefaultGasValue = uint64(17179869184)
)

var AggregateMethodBytes []byte

func New(_client royal.EthClient, opts ...Option) (Multicall, error) {
    // Initialize the parsing of the hex bytes into a global var
    if AggregateMethodBytes == nil {
        aggMethodBytes, err := hexutil.Decode(AggregateMethod)
        if err != nil {
            return nil, fmt.Errorf("unable to decode the aggregate method string: (err: %s)\n", err)
        }

        AggregateMethodBytes = aggMethodBytes
    }

    config := &Config{
        Gas: DefaultGasValue,
    }

    for _, opt := range opts {
        opt(config)
    }

    return &multicall{
        config: config,
        client: _client,
    }, nil
}

type CallResult struct {
    Raw     []byte
    Decoded []interface{}
}

type Result struct {
    BlockNumber uint64
    Calls       map[string]CallResult
}

func (mc multicall) CallRaw(calls ViewCalls, block string) (*Result, error) {
    resultRaw, err := mc.makeRequest(calls, block)
    if err != nil {
        return nil, err
    }
    return calls.decodeRaw(resultRaw)
}

func (mc multicall) Call(calls ViewCalls, block string) (*Result, error) {
    // Filter out empty calls and return a sane response
    if len(calls) == 0 {
        blockNum, err := hexutil.DecodeUint64(block)
        if err != nil {
            return nil, fmt.Errorf("unknown block number; must be hex: %s\n", err)
        }

        calls := make(map[string]CallResult)
        return &Result{
            BlockNumber: blockNum,
            Calls:       calls,
        }, nil
    }

    // if we are actually making a call, go ahead and proxy it
    resultRaw, err := mc.makeRequest(calls, block)
    if err != nil {
        return nil, err
    }
    return calls.decode(resultRaw)
}

func (mc multicall) makeRequest(calls ViewCalls, block string) (string, error) {
    to := common.HexToAddress(mc.config.MulticallAddress)

    // Convert the calls into abi packed bytes for transmission to the chain
    payloadArgs, err := calls.callData()
    if err != nil {
        return "", fmt.Errorf("unable to serialize calldata (err: %s)\n", err)
    }

    // TODO: This should be done elsewhere
    // Pre-pend the method bytes to the beginning of the payload
    payloadArgs = append(AggregateMethodBytes, payloadArgs...)

    callMsg := ethereum.CallMsg{
        To:   &to,
        Gas:  mc.config.Gas,
        Data: payloadArgs,
    }

    // TODO: Require a *big.Int blockNumber
    // Parse block number from the input string
    blockBig, ok := new(big.Int).SetString(block, 10)
    if !ok {
        blockBig = nil
    }

    resultBytes, err := mc.client.CallContract(context.Background(), callMsg, blockBig)
    if err != nil {
        return "", fmt.Errorf("trouble issuing eth_call for multicall: (err: %s)\n", err)
    }

    // TODO: Change the interface between these methods to be []byte instead of string
    // Return a hex-encoded string
    return hexutil.Encode(resultBytes), nil
}

func (mc multicall) Contract() string {
    return mc.config.MulticallAddress
}
