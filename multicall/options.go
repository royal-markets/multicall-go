package multicall

type Option func(*Config)

// Config defines the Config for a multicall client
type Config struct {
    MulticallAddress string
    Gas              uint64
}

const (
// TODO: Add Multicall addresses for various networks
)

// SetContractAddress will update the contract address for the Config
func SetContractAddress(address string) Option {
    return func(c *Config) {
        c.MulticallAddress = address
    }
}

// SetGas will update the gas limit to the given value
func SetGas(gas uint64) Option {
    return func(c *Config) {
        c.Gas = gas
    }
}
