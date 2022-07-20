package multicall

import (
    "context"
    "fmt"

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

func New(_client royal.EthClient, opts ...Option) (Multicall, error) {
    config := &Config{
        MulticallAddress: "",
        Gas:              17179869184,
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
    //Success bool
    Raw     []byte
    Decoded []interface{}
}

type Result struct {
    BlockNumber uint64
    Calls       map[string]CallResult
}

const AggregateMethod = "0x252dba42"

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

    // TODO: Do this once instead of many times
    aggMethodBytes, err := hexutil.Decode(AggregateMethod)
    if err != nil {
        return "", fmt.Errorf("unable to decode the aggregate method string: (err: %s)\n", err)
    }

    // Convert the calls into abi packed bytes for transmission to the chain
    payloadArgs, err := calls.callData()
    if err != nil {
        return "", fmt.Errorf("unable to serialize calldata (err: %s)\n", err)
    }

    // Pre-pend the method bytes to the beginning of the payload
    // TODO: This should be done elsewhere
    payloadArgs = append(aggMethodBytes, payloadArgs...)

    callMsg := ethereum.CallMsg{
        To:   &to,
        Gas:  mc.config.Gas,
        Data: payloadArgs,
    }

    // TODO: Actually use the blockNumber
    resultBytes, err := mc.client.CallContract(context.Background(), callMsg, nil)
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
