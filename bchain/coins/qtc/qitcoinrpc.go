package qtc

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
	"github.com/trezor/blockbook/common"
)

// QitcoinRPC is an interface to JSON-RPC bitcoind service
type QitcoinRPC struct {
	*btc.BitcoinRPC
}

type ResGetBlockChainInfo struct {
	Error  *bchain.RPCError `json:"error"`
	Result struct {
		Chain         string            `json:"chain"`
		Blocks        int               `json:"blocks"`
		Headers       int               `json:"headers"`
		Bestblockhash string            `json:"bestblockhash"`
		Difficulty    common.JSONNumber `json:"difficulty"`
		Pruned        bool              `json:"pruned"`
		SizeOnDisk    int64             `json:"size_on_disk"`
		Consensus     struct {
			Chaintip  string `json:"chaintip"`
			Nextblock string `json:"nextblock"`
		} `json:"consensus"`
	} `json:"result"`
}

// NewQitcoinRPC returns new QitcoinRPC instance
func NewQitcoinRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}
	q := &QitcoinRPC{
		BitcoinRPC: b.(*btc.BitcoinRPC),
	}
	q.RPCMarshaler = btc.JSONMarshalerV2{}
	q.ChainConfig.SupportsEstimateSmartFee = false
	return q, nil
}

// Initialize initializes QitcoinRPC instance
func (q *QitcoinRPC) Initialize() error {
	ci, err := q.GetChainInfo()
	if err != nil {
		return err
	}
	chainName := ci.Chain

	params := GetChainParams(chainName)

	q.Parser = NewQitcoinParser(params, q.ChainConfig)

	// parameters for getInfo request
	if params.Net == MainnetMagic {
		q.Testnet = false
		q.Network = "livenet"
	} else {
		q.Testnet = true
		q.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)

	return nil
}

func (q *QitcoinRPC) GetChainInfo() (*bchain.ChainInfo, error) {
	chainInfo := ResGetBlockChainInfo{}
	err := q.Call(&btc.CmdGetBlockChainInfo{Method: "getblockchaininfo"}, &chainInfo)
	if err != nil {
		return nil, err
	}
	if chainInfo.Error != nil {
		return nil, chainInfo.Error
	}

	networkInfo := btc.ResGetNetworkInfo{}
	err = q.Call(&btc.CmdGetNetworkInfo{Method: "getnetworkinfo"}, &networkInfo)
	if err != nil {
		return nil, err
	}
	if networkInfo.Error != nil {
		return nil, networkInfo.Error
	}

	return &bchain.ChainInfo{
		Bestblockhash:   chainInfo.Result.Bestblockhash,
		Blocks:          chainInfo.Result.Blocks,
		Chain:           chainInfo.Result.Chain,
		Difficulty:      string(chainInfo.Result.Difficulty),
		Headers:         chainInfo.Result.Headers,
		SizeOnDisk:      chainInfo.Result.SizeOnDisk,
		Version:         string(networkInfo.Result.Version),
		Subversion:      string(networkInfo.Result.Subversion),
		ProtocolVersion: string(networkInfo.Result.ProtocolVersion),
		Timeoffset:      networkInfo.Result.Timeoffset,
		Consensus:       chainInfo.Result.Consensus,
		Warnings:        networkInfo.Result.Warnings,
	}, nil
}

// GetBlock returns block with given hash.
func (q *QitcoinRPC) GetBlock(hash string, height uint32) (*bchain.Block, error) {
	var err error
	if hash == "" {
		hash, err = q.GetBlockHash(height)
		if err != nil {
			return nil, err
		}
	}
	if !q.ParseBlocks {
		return q.GetBlockFull(hash)
	}
	// optimization
	if height > 0 {
		return q.GetBlockWithoutHeader(hash, height)
	}
	header, err := q.GetBlockHeader(hash)
	if err != nil {
		return nil, err
	}
	data, err := q.GetBlockRaw(hash)
	if err != nil {
		return nil, err
	}
	block, err := q.Parser.ParseBlock(data)
	if err != nil {
		return nil, errors.Annotatef(err, "hash %v", hash)
	}
	block.BlockHeader = *header
	return block, nil
}

// GetTransactionForMempool returns a transaction by the transaction ID.
// It could be optimized for mempool, i.e. without block time and confirmations
func (q *QitcoinRPC) GetTransactionForMempool(txid string) (*bchain.Tx, error) {
	return q.GetTransaction(txid)
}

// GetMempoolEntry returns mempool data for given transaction
func (q *QitcoinRPC) GetMempoolEntry(txid string) (*bchain.MempoolEntry, error) {
	return nil, errors.New("GetMempoolEntry: not implemented")
}

func isErrBlockNotFound(err *bchain.RPCError) bool {
	return err.Message == "Block not found" ||
		err.Message == "Block height out of range"
}
