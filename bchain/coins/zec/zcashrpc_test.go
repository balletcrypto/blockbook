package zec

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestZcashGetBlock(t *testing.T) {
	configfile := "/Users/johnsoncai/.go/src/blockbook/zcash-config.json"
	data, err := ioutil.ReadFile(configfile)
	if err != nil {
		t.Fatal(err)
	}
	var config json.RawMessage
	err = json.Unmarshal(data, &config)
	if err != nil {
		t.Fatal(err)
	}
	rpc, err := NewZCashRPC(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	block, err := rpc.GetBlock("00000000000b6e22763c5fd3aa89bdbf331b282a8d077c9cd612127bc2233acb", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(block.Confirmations)
}
