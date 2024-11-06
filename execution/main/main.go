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
	// "golang.org/x/text/cases"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpc := "https://eth-mainnet.g.alchemy.com/v2/j28GcevSYukh-GvSeBOYcwHOfIggF1Gt"
	blockChan := make(chan *common.Block, 100)
	finalizedBlockChan := make(chan *common.Block, 100)

	state := execution.NewState(64, blockChan, finalizedBlockChan)
	var executionClient *execution.ExecutionClient
	executionClient, _ = executionClient.New(rpc, state)

	go func() {
		UpdateState(ctx, cancel, state)
	}()

	time.Sleep(10 * time.Second)
	block, err := executionClient.GetBlock(common.BlockTag{Number: 21123701}, false)
	if err != nil {
		fmt.Printf("Error in executionGetBlock: %v\n", err)
	}
	fmt.Printf("Block: %v\n", block)

	// Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // Cleanup
    cancel()  // Signal all goroutines to stop
}

func UpdateState(ctx context.Context, cancel context.CancelFunc, state *execution.State) {
	// ctx, cancel := context.WithCancel(context.Background())
	rpc := "https://eth-mainnet.g.alchemy.com/v2/j28GcevSYukh-GvSeBOYcwHOfIggF1Gt"
	defer cancel()

	
	lastFetchedBlock := uint64(21123700)

	rpcClient, err := execution.NewRPCClient(rpc)
	if err != nil {
		fmt.Printf("error in creating execution client: %v", err)
	}
	
	newBlockChan := make(chan *common.Block, 100)
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
				_, err := execution.FetchBlocksFromRPC(ctx, rpcClient, lastFetchedBlock, newBlockChan)
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
			execution.UpdateState(state, newBlockChan)
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)

			fmt.Println("Most recent block in state: ", state.LatestBlockNumber())
		}
	}()

	// Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // Cleanup
    cancel()  // Signal all goroutines to stop
}