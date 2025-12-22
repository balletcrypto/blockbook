package xec

import (
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/martinboehm/btcutil/txscript"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
)

const (
	// MainNetPrefix is CashAddr prefix for mainnet
	MainNetPrefix = "ecash:"
	// TestNetPrefix is CashAddr prefix for testnet
	TestNetPrefix = "ectest:"
	// RegTestPrefix is CashAddr prefix for regtest
	RegTestPrefix = "ecregtest:"
)

type EcashParser struct {
	*btc.BitcoinParser
}

func NewEcashParser(params *chaincfg.Params, c *btc.Configuration) (*EcashParser, error) {
	p := btc.NewBitcoinParser(params, c)
	p.BaseParser.AmountDecimalPoint = 2

	e := &EcashParser{BitcoinParser: p}
	e.OutputScriptToAddressesFunc = e.outputScriptToAddresses
	return e, nil
}

func GetChainParams(chain string) *chaincfg.Params {
	if !chaincfg.IsRegistered(&MainNetParams) {
		err := chaincfg.Register(&MainNetParams)
		if err == nil {
			err = chaincfg.Register(&TestNetParams)
		}
		if err == nil {
			err = chaincfg.Register(&RegtestParams)
		}
		if err != nil {
			panic(err)
		}
	}
	switch chain {
	case "test":
		return &TestNetParams
	case "regtest":
		return &RegtestParams
	default:
		return &MainNetParams
	}
}

// GetAddrDescFromAddress returns internal address representation of given address
func (p *EcashParser) GetAddrDescFromAddress(address string) (bchain.AddressDescriptor, error) {
	return p.addressToOutputScript(address)
}

// addressToOutputScript converts bitcoin address to ScriptPubKey
func (p *EcashParser) addressToOutputScript(address string) ([]byte, error) {
	if isCashAddr(address) {
		da, err := DecodeAddress(address, p.Params)
		if err != nil {
			return nil, err
		}
		script, err := PayToAddrScript(da)
		if err != nil {
			return nil, err
		}
		return script, nil
	}
	da, err := DecodeAddress(address, p.Params)
	if err != nil {
		return nil, err
	}
	script, err := txscript.PayToAddrScript(da)
	if err != nil {
		return nil, err
	}
	return script, nil
}

func isCashAddr(addr string) bool {
	n := len(addr)
	switch {
	case n > len(MainNetPrefix) && addr[0:len(MainNetPrefix)] == MainNetPrefix:
		return true
	case n > len(TestNetPrefix) && addr[0:len(TestNetPrefix)] == TestNetPrefix:
		return true
	case n > len(RegTestPrefix) && addr[0:len(RegTestPrefix)] == RegTestPrefix:
		return true
	}

	return false
}

// outputScriptToAddresses converts ScriptPubKey to bitcoin addresses
func (p *EcashParser) outputScriptToAddresses(script []byte) ([]string, bool, error) {
	// convert possible P2PK script to P2PK, which bchutil can process
	var err error
	script, err = txscript.ConvertP2PKtoP2PKH(p.Params.Base58CksumHasher, script)
	if err != nil {
		return nil, false, err
	}
	a, err := ExtractPkScriptAddrs(script, p.Params)
	if err != nil {
		// do not return unknown script type error as error
		if err.Error() == "unknown script type" {
			// try OP_RETURN script
			or := p.TryParseOPReturn(script)
			if or != "" {
				return []string{or}, false, nil
			}
			return []string{}, false, nil
		}
		return nil, false, err
	}
	// EncodeAddress returns CashAddr address
	addr := a.EncodeAddress()
	return []string{addr}, len(addr) > 0, nil
}
