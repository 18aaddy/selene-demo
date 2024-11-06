package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/BlocSoc-iitr/selene/common"
	"github.com/BlocSoc-iitr/selene/config"
	"github.com/BlocSoc-iitr/selene/execution"
	Common "github.com/ethereum/go-ethereum/common"
)

func TestGetAccount(t *testing.T) {
	var n config.Network
	var client Client
	blockChan := make(chan *common.Block, 100)
	newBlockChan := make(chan *common.Block, 100)
	finalizedBlockChan := make(chan *common.Block, 100)

	state := execution.NewState(64, blockChan, finalizedBlockChan)

	go func(){
		execution.UpdateState(state, newBlockChan)
	}()

	baseConfig, err := n.BaseConfig("MAINNET")
	if err != nil {
		fmt.Printf("Error in base config creation: %v", err)
	}

	checkpoint := "b21924031f38635d45297d68e7b7a408d40b194d435b25eeccad41c522841bd5"
	consensusRpcUrl := "http://testing.mainnet.beacon-api.nimbus.team"
	executionRpcUrl := "https://eth-mainnet.g.alchemy.com/v2/j28GcevSYukh-GvSeBOYcwHOfIggF1Gt"
	rpcBindIp := "127.0.0.1"
	rpcPort := uint16(8545)

	config := config.Config{
		ConsensusRpc:         consensusRpcUrl,
		ExecutionRpc:         executionRpcUrl,
		RpcBindIp:            &rpcBindIp,
		RpcPort:              &rpcPort,
		DefaultCheckpoint:    baseConfig.DefaultCheckpoint,
		Checkpoint:           (*[32]byte)(Common.Hex2Bytes(checkpoint)),
		Chain:                baseConfig.Chain,
		Forks:                baseConfig.Forks,
		StrictCheckpointAge:  false,
		DataDir:              baseConfig.DataDir,
		DatabaseType:         nil,
		MaxCheckpointAge:     baseConfig.MaxCheckpointAge,
		LoadExternalFallback: false,
	}

	client, err = client.New(config, state)

	go func() {
		time.Sleep(20 * time.Second)
		address := common.Address{Addr: [20]byte(Common.Hex2Bytes("62d4d3785f8117be8d2ee8e1e81c9147098bc3ff"))}
		tag:= common.BlockTag{Latest: true}
		if client.Node.Execution == nil {
			t.Error("ExecutionClient is nil")
		}
		chain, _ := client.Node.Execution.Rpc.ChainId()
		fmt.Printf("Rpc: %d", chain)
		block,err := client.Node.Execution.GetAccount(&address, &[]Common.Hash{}, tag)
		if err != nil {
			t.Errorf("Error: %v", err)
		} else {
			t.Logf("No error: %v", block)
		}
		// fmt.Println("Account: ", acc)
		t.Errorf("ReachedEnd")
	}()

	select{}
}