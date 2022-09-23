package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

const key = `json object from keystore`

func main() {
	PrivateKey, _ := crypto.HexToECDSA("your privatekey")

	// Create an IPC based RPC connection to a remote node and instantiate a contract binding
	conn, err := ethclient.Dial("https://goerli.infura.io/v3/980eaff455ba497aa86a0d192ea862a8")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	store, err := NewStorage(common.HexToAddress("contract address"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Storage contract: %v", err)
	}
	// Create an authorized transactor and call the store function
	nonce, _ := conn.NonceAt(context.Background(), common.HexToAddress("your address"), nil)
	gasPrice, _ := conn.SuggestGasPrice(context.Background())
	auth, err := bind.NewKeyedTransactorWithChainID(PrivateKey, big.NewInt(5))
	auth.GasLimit = uint64(300000)
	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.GasPrice = gasPrice

	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	// Call the store() function
	tx, err := store.Store(auth, big.NewInt(420))
	if err != nil {
		log.Fatalf("Failed to update value: %v", err)
	}
	fmt.Printf("Update pending: 0x%x\n", tx.Hash())

}
