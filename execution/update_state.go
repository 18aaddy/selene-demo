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
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
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

	var tempBlock struct {
		Number          hexutil.Uint64
		BaseFeePerGas   hexutil.Big
		Difficulty      hexutil.Big
		ExtraData       hexutil.Bytes
		GasLimit        hexutil.Uint64
		GasUsed         hexutil.Uint64
		Hash            hexutil.Bytes
		LogsBloom       hexutil.Bytes
		Miner           string
		MixHash         hexutil.Bytes
		Nonce           string
		ParentHash      hexutil.Bytes
		ReceiptsRoot    hexutil.Bytes
		Sha3Uncles      hexutil.Bytes
		Size            hexutil.Uint64
		StateRoot       hexutil.Bytes
		Timestamp       hexutil.Uint64
		TotalDifficulty hexutil.Big
		// Transactions    struct {
		// 	Hashes []hexutil.Bytes
		// 	Full   []struct {
		// 		AccessList           AccessList      `json:"accessList"`
		// 		Hash                 hexutil.Bytes   `json:"hash"`
		// 		Nonce                hexutil.Uint64  `json:"nonce"`
		// 		BlockHash            hexutil.Bytes   `json:"blockHash"`   // Pointer because it's nullable
		// 		BlockNumber          hexutil.Uint64  `json:"blockNumber"` // Pointer because it's nullable
		// 		TransactionIndex     hexutil.Uint64  `json:"transactionIndex"`
		// 		From                 *hexutil.Bytes  `json:"from"`
		// 		To                   *hexutil.Bytes  `json:"to"` // Pointer because 'to' can be null for contract creation
		// 		Value                hexutil.Big     `json:"value"`
		// 		GasPrice             hexutil.Big     `json:"gasPrice"`
		// 		Gas                  hexutil.Uint64  `json:"gas"`
		// 		Input                hexutil.Bytes   `json:"input"`
		// 		ChainID              hexutil.Big     `json:"chainId"`
		// 		TransactionType      hexutil.Uint    `json:"type"`
		// 		Signature            *common.Signature      `json:"signature"`
		// 		MaxFeePerGas         hexutil.Big     `json:"maxFeePerGas"`
		// 		MaxPriorityFeePerGas hexutil.Big     `json:"maxPriorityFeePerGas"`
		// 		MaxFeePerBlobGas     hexutil.Big     `json:"maxFeePerBlobGas"`
		// 		BlobVersionedHashes  []hexutil.Bytes `json:"blobVersionedHashes"`
		// 	}
		// }
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
    // block.BaseFeePerGas = tempBlock.BaseFeePerGas
    // block.Difficulty = tempBlock.Difficulty
    block.ExtraData = tempBlock.ExtraData
    block.GasLimit = uint64(tempBlock.GasLimit)

		// GasUsed         hexutil.Uint64
		// Hash            hexutil.Bytes
		// LogsBloom       hexutil.Bytes
		// Miner           string
		// MixHash         hexutil.Bytes
		// Nonce           string
		// ParentHash      hexutil.Bytes
		// ReceiptsRoot    hexutil.Bytes
		// Sha3Uncles      hexutil.Bytes
		// Size            hexutil.Uint64
		// StateRoot       hexutil.Bytes
		// Timestamp       hexutil.Uint64
		// TotalDifficulty hexutil.Uint64

	return &block, nil
}
