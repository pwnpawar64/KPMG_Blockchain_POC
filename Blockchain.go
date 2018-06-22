package main

import (
	"fmt"

	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/rs/xid"
	"strconv"
	"time"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

type Transaction struct {
	TransactionId        xid.ID    `json:"txId"`
	TransactionTimestamp time.Time `json:"txTimestamp"`
	TransactionType      string    `json:"txType"`
	TransactionOwner     string    `json:"txOwner"`
	ProductInfo          Product   `json:"productInfo"`
}

type Product struct {
	RetailerId  int    `json:"retailerId"`
	SupplierId  int    `json:"supplierId"`
	ProductId   int    `json:"productId"`
	ProductName string `json:"productName"`
	Brand       string `json:"brand"`
	Style       string `json:"style"`
	Size        int    `json:"size"`
	Color       string `json:"color"`
	Quantity    int    `json:"quantity"`
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	//args := stub.GetStringArgs()

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	if fn == "addInventory" {
		fmt.Println(args)
		return (addInventory(stub, args))
	} else if fn == "viewInventory" {
		return (viewInventory(stub, args))
	} else if fn == "sellFromInventory" {
		return (sellFromInventory(stub, args))
	} else if fn == "getTransactionHistory" {
		return (getTransactionHistory(stub, args))
	}
	return shim.Error("Invalid Smart Contract function name.")

}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func addInventory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 9 {
		return shim.Error("Incorrect arguments. Expecting 9")
	}
	var product Product
	product.RetailerId, _ = strconv.Atoi(args[0])
	product.SupplierId, _ = strconv.Atoi(args[1])
	product.ProductId, _ = strconv.Atoi(args[2])
	product.ProductName = args[3]
	product.Brand = args[4]
	product.Style = args[5]
	product.Size, _ = strconv.Atoi(args[6])
	product.Color = args[7]
	product.Quantity, _ = strconv.Atoi(args[8])

	var transaction Transaction
	transaction.TransactionId = xid.New()
	transaction.TransactionTimestamp = time.Now()
	transaction.TransactionType = "add"
	transaction.TransactionOwner = args[1]
	transaction.ProductInfo = product

	transactionAsBytes, err := json.Marshal(transaction)
	if err != nil {
		return shim.Error("Something went wrong in chaincode")
	}
	productAsBytes, err := json.Marshal(product)
	if err != nil {
		return shim.Error("Something went wrong in chaincode")
	}
	//First update the transaction
	err = stub.PutState(strconv.Itoa(product.RetailerId), transactionAsBytes)
	if err != nil {
		return shim.Error("Failed to set transaction ")
	}
	//TO DO: Check if product is already present
	//Update the product
	err = stub.PutState(args[2], productAsBytes)
	if err != nil {

		// If updating product document fails, remove the transaction also
		//stub.DelState(transaction.TransactionId.String())
		return shim.Error("Failed to set asset: " + args[2])
	}
	return shim.Success(productAsBytes)
}

// View the latest quantity of all products
func viewInventory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get asset: %s with error: " + args[0])
	}
	if value == nil {
		return shim.Error("Asset not found: " + args[0])
	}
	return shim.Success(value)
}

func sellFromInventory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting 2")
	}
	productAsBytes, _ := stub.GetState(args[0])
	if productAsBytes==nil {
		return shim.Error("Product not present in inventory")
	}
	product := Product{}
	json.Unmarshal(productAsBytes, &product)

	qtyToSell,_:=strconv.Atoi(args[1])
	if product.Quantity < qtyToSell  {
		return shim.Error("Inventory not sufficient")
	}
	product.Quantity = product.Quantity-qtyToSell

	var transaction Transaction
	transaction.TransactionId = xid.New()
	transaction.TransactionTimestamp = time.Now()
	transaction.TransactionType = "sell"
	transaction.TransactionOwner = strconv.Itoa(product.RetailerId)
	transaction.ProductInfo = product

	transactionAsBytes, err := json.Marshal(transaction)
	if err != nil {
		return shim.Error("Something went wrong in chaincode")
	}

	//First update the transaction
	err = stub.PutState(strconv.Itoa(product.RetailerId), transactionAsBytes)
	if err != nil {
		return shim.Error("Failed to set transaction ")
	}

	productAsBytes, _ = json.Marshal(product)
	err = stub.PutState(args[0], productAsBytes)
	if err != nil {
		return shim.Error("Something went wrong in chaincode")
	}
	return shim.Success(productAsBytes)
}

func getTransactionHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get asset: %s with error: " + args[0])
	}
	if value == nil {
		return shim.Error("Asset not found: " + args[0])
	}
	return shim.Success(value)
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
