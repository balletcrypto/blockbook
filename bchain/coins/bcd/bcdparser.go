package bcd

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/bsm/go-vlq"
	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
	"io"
)

type BitcoinDiamondParser struct {
	*btc.BitcoinParser
}

func NewBitcoinDiamondParser(params *chaincfg.Params, c *btc.Configuration) *BitcoinDiamondParser {
	p := btc.NewBitcoinParser(params, c)
	p.BaseParser.AmountDecimalPoint = 7
	return &BitcoinDiamondParser{BitcoinParser: p}
}

func (p *BitcoinDiamondParser) ParseTx(b []byte) (*bchain.Tx, error) {
	t := wire.MsgTx{}
	r := bytes.NewReader(b)

	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	version := binary.LittleEndian.Uint32(buf)
	if version == 12 {
		r.Seek(32, 0)
	} else {
		r.Seek(0, 0)
	}

	if err := t.Deserialize(r); err != nil {
		return nil, err
	}
	tx := p.TxFromMsgTx(&t, true)
	tx.Hex = hex.EncodeToString(b)
	tx.Version = int32(version)
	return &tx, nil
}

func (p *BitcoinDiamondParser) UnpackTx(buf []byte) (*bchain.Tx, uint32, error) {
	height := binary.BigEndian.Uint32(buf)
	bt, l := vlq.Int(buf[4:])
	tx, err := p.ParseTx(buf[4+l:])
	if err != nil {
		return nil, 0, err
	}
	tx.Blocktime = bt

	return tx, height, nil
}
