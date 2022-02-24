package xec

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/martinboehm/bchutil"
	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
)

var (
	// MainNetParams are parser parameters for mainnet
	MainNetParams chaincfg.Params
	// TestNetParams are parser parameters for testnet
	TestNetParams chaincfg.Params
	// RegtestParams are parser parameters for regtest
	RegtestParams chaincfg.Params
)

func init() {
	MainNetParams = chaincfg.MainNetParams
	MainNetParams.Net = bchutil.MainnetMagic

	TestNetParams = chaincfg.TestNet3Params
	TestNetParams.Net = bchutil.TestnetMagic

	RegtestParams = chaincfg.RegressionNetParams
	RegtestParams.Net = bchutil.Regtestmagic
}

// BGoldRPC is an interface to JSON-RPC bitcoind service.
type EcashRPC struct {
	*btc.BitcoinRPC
}

// NewBGoldRPC returns new BGoldRPC instance.
func NewEcashRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}

	s := &EcashRPC{
		b.(*btc.BitcoinRPC),
	}

	return s, nil
}

// Initialize initializes BGoldRPC instance.
func (b *EcashRPC) Initialize() error {
	ci, err := b.GetChainInfo()
	if err != nil {
		return err
	}
	chainName := ci.Chain

	params := GetChainParams(chainName)

	// always create parser
	b.Parser, err = NewEcashParser(params, b.ChainConfig)

	if err != nil {
		return err
	}

	// parameters for getInfo request
	if params.Net == wire.MainNet {
		b.Testnet = false
		b.Network = "livenet"
	} else {
		b.Testnet = true
		b.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)
	return nil
}
