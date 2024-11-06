package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BlocSoc-iitr/selene/common"
	"github.com/BlocSoc-iitr/selene/execution"
	"github.com/BlocSoc-iitr/selene/client"
	Common "github.com/ethereum/go-ethereum/common"
	"github.com/BlocSoc-iitr/selene/config"
	// "golang.org/x/text/cases"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var client client.Client

	// Create a channel to receive OS signals.
    sigs := make(chan os.Signal, 1)
    // Notify the channel on interrupt signals.
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    // Run the shutdown process in a goroutine when signal is received.
    go func() {
        <-sigs
        client.Shutdown()
		cancel()  // Signal all goroutines to stop
        os.Exit(0)
    }()

	blockChan := make(chan *common.Block, 100)
	finalizedBlockChan := make(chan *common.Block, 100)

	state := execution.NewState(64, blockChan, finalizedBlockChan)
	// var executionClient *execution.ExecutionClient
	// executionClient, _ = executionClient.New(rpc, state)
	// config := config.Config{}
	
	go func() {
		UpdateState(ctx, cancel, state, blockChan)
	}()

	time.Sleep(10 * time.Second)

	var n config.Network
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
	
	if client.Node == nil{
		fmt.Println("Did not expect node to be nil")
	}
	if client.Rpc == nil{
		fmt.Println("Did not expect node to be nil")
	}

	if err != nil {
		fmt.Printf("Error in client creation: %v", err)
	}

	err = client.Start()
	if err != nil {
		fmt.Printf("Error in starting client: %v", err)
	}

	if err != nil {
		fmt.Printf("Error in newClient: %v\n", err)
	}


	// // block, err := executionClient.GetBlock(common.BlockTag{Number: 21123701}, false)
	
	// if err != nil {
	// 	fmt.Printf("Error in executionGetBlock: %v\n", err)
	// }
	// // fmt.Printf("Block: %v\n", block)

	// // Wait for interrupt signal
    // sigChan := make(chan os.Signal, 1)
    // signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    // <-sigChan

    // Cleanup
	select{}
}

func UpdateState(ctx context.Context, cancel context.CancelFunc, state *execution.State, blockChan chan *common.Block) {
	// ctx, cancel := context.WithCancel(context.Background())
	rpc := "https://eth-mainnet.g.alchemy.com/v2/j28GcevSYukh-GvSeBOYcwHOfIggF1Gt"
	defer cancel()

	
	lastFetchedBlock := uint64(21123700)

	rpcClient, err := execution.NewRPCClient(rpc)
	if err != nil {
		fmt.Printf("error in creating execution client: %v", err)
	}
	
	// newBlockChan := make(chan *common.Block, 100)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				latestBlockNumber := state.LatestBlockNumber()
				if latestBlockNumber != nil {
					lastFetchedBlock = *latestBlockNumber
				}
				_, err := execution.FetchBlocksFromRPC(ctx, rpcClient, lastFetchedBlock, blockChan)
				if err != nil {
					// Handle error
					fmt.Printf("Error: %v", err)
					continue
				}
				// fmt.Println("Reached here @$")
			}
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)	
			execution.UpdateState(state, blockChan)
		}
	}()

	go func() {
		for {
			time.Sleep(10 * time.Second)

			fmt.Println("Most recent block in state: ", *(state.LatestBlockNumber()))
			// block := state.GetBlock(common.BlockTag{Number: 21123710})
			fmt.Println("Block in state: ", state)
		}
	}()

	// Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // Cleanup
    cancel()  // Signal all goroutines to stop
}