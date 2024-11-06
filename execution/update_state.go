package execution

import (
	"context"
	// "encoding/json"
	"fmt"
	"math/big"

	// "math/big"
	// "strconv"
	// "strings"
	"time"

	"github.com/BlocSoc-iitr/selene/common"
	"github.com/BlocSoc-iitr/selene/utils"

	// Common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
	// "github.com/holiman/uint256"
)

func UpdateState(s *State, blockChan chan *common.Block) {
	fmt.Println("Reached UpdateState")
	for newBlock := range blockChan {
		if newBlock == nil {
			continue
		}

		// s.mu.Lock()
		// fmt.Printf("Processing block number: %v\n", newBlock.Number)
		s.PushBlock(newBlock)
		// fmt.Println("State block map: ", s.blocks)
		// s.mu.Unlock()
	}
}

func FetchBlocksFromRPC(ctx context.Context, rpcClient *RPCClient, lastFetchedBlock uint64, blockChan chan *common.Block) ([]*common.Block, error) {
	var newBlocks []*common.Block
	fmt.Println("Reached FetchBlocksFromRPC")

	for {
		select {
		case <-ctx.Done():
			return newBlocks, ctx.Err()
		case <-time.After(5 * time.Second):
			// Fetch new blocks from the RPC
			blocks, err := fetchNewBlocks(rpcClient, lastFetchedBlock, blockChan)
			if err != nil {
				return newBlocks, err
			}

			newBlocks = append(newBlocks, blocks...)
			if len(blocks) > 0 {
				lastFetchedBlock = blocks[len(blocks)-1].Number
				fmt.Printf("Last Fetched Block Number: %d", lastFetchedBlock)
			}
		}
	}
}

func fetchNewBlocks(rpcClient *RPCClient, lastFetchedBlock uint64, blockChan chan *common.Block) ([]*common.Block, error) {
	var blocks []*common.Block
	fmt.Println("Reached fetchNewBlocks")
	// Fetch blocks from the RPC
	for i := lastFetchedBlock + 1; ; i++ {
		block, err := rpcClient.GetBlockByNumber(common.BlockTag{Number: i}.String(), true)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return blocks, err
		}
		// fmt.Printf("Sending block %v through channel\n", block.Number)
		blockChan <- block
		if block.Number == 0 {
			break
		}

		blocks = append(blocks, block)
	}

	close(blockChan)

	return blocks, nil
}

// RPCClient wraps the ethereum rpc client
type RPCClient struct {
	client *rpc.Client
}

// NewRPCClient creates a new RPC client
func NewRPCClient(rpcURL string) (*RPCClient, error) {
	// Create a new client with default options
	client, err := rpc.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	return &RPCClient{
		client: client,
	}, nil
}

// GetBlockByNumber fetches a block by its number or tag
func (c *RPCClient) GetBlockByNumber(blockTag string, fullTx bool) (*common.Block, error) {
	var block common.Block
	var blockID interface{}
	type Signature struct {
		R hexutil.Big `json:"r"`
		S hexutil.Big `json:"s"`
		V hexutil.Big `json:"v"`
	}

	var tempBlock struct {
		Number           hexutil.Uint64
		BaseFeePerGas    hexutil.Big
		Difficulty       hexutil.Big
		ExtraData        hexutil.Bytes
		GasLimit         hexutil.Uint64
		GasUsed          hexutil.Uint64
		Hash             hexutil.Bytes
		LogsBloom        hexutil.Bytes
		Miner            string
		MixHash          hexutil.Bytes
		Nonce            string
		ParentHash       hexutil.Bytes
		ReceiptsRoot     hexutil.Bytes
		Sha3Uncles       hexutil.Bytes
		Size             hexutil.Uint64
		StateRoot        hexutil.Bytes
		Timestamp        hexutil.Uint64
		TotalDifficulty  hexutil.Big
		Transactions     interface{}
		TransactionsRoot hexutil.Bytes
		Uncles           []hexutil.Bytes
		BlobGasUsed      *hexutil.Uint64
		ExcessBlobGas    *hexutil.Uint64
	}

	// Handle block tag
	switch blockTag {
	case "latest", "pending", "finalized", "safe":
		blockID = blockTag
	default:
		// Try to parse as hex if it starts with 0x
		if len(blockTag) > 2 && blockTag[:2] == "0x" {
			blockID = blockTag
		} else {
			// Parse as decimal number
			blockNum, ok := new(big.Int).SetString(blockTag, 10)
			if !ok {
				return nil, fmt.Errorf("invalid block tag: %s", blockTag)
			}
			blockID = hexutil.EncodeBig(blockNum)
		}
	}

	err := c.client.Call(&tempBlock, "eth_getBlockByNumber", blockID, fullTx)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}

	block.Number = uint64(tempBlock.Number)
	block.BaseFeePerGas = *uint256.MustFromBig(tempBlock.BaseFeePerGas.ToInt())
	block.Difficulty = *uint256.MustFromBig(tempBlock.Difficulty.ToInt())
	block.ExtraData = tempBlock.ExtraData
	block.GasLimit = uint64(tempBlock.GasLimit)
	block.GasUsed = uint64(tempBlock.GasUsed)
	block.Hash = [32]byte(tempBlock.Hash)
	block.LogsBloom = tempBlock.LogsBloom
	addr, _ := utils.Hex_str_to_bytes(tempBlock.Miner)
	block.Miner = common.Address{Addr: [20]byte(addr)}
	block.MixHash = [32]byte(tempBlock.MixHash)
	block.Nonce = tempBlock.Nonce
	block.ParentHash = [32]byte(tempBlock.ParentHash)
	block.ReceiptsRoot = [32]byte(tempBlock.ReceiptsRoot)
	block.Sha3Uncles = [32]byte(tempBlock.Sha3Uncles)
	block.Size = uint64(tempBlock.Size)
	block.StateRoot = [32]byte(tempBlock.StateRoot)
	block.Timestamp = uint64(tempBlock.Timestamp)
	block.TotalDifficulty = *uint256.MustFromBig(tempBlock.TotalDifficulty.ToInt())
	block.TransactionsRoot = [32]byte(tempBlock.TransactionsRoot)
	for _, uncle := range tempBlock.Uncles {
		block.Uncles = append(block.Uncles, [32]byte(uncle))
	}
	block.BlobGasUsed = (*uint64)(tempBlock.BlobGasUsed)
	block.ExcessBlobGas = (*uint64)(tempBlock.ExcessBlobGas)

	return &block, nil
}
