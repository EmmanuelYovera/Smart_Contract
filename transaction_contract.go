package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type RealEstateTransaction struct {
	contractapi.Contract
}

type Transaction struct {
	ID                 int    `json:"id"`
	PartidaRegistral   string `json:"partidaRegistral"`
	CompradorDNI       string `json:"compradorDni"`
	CompradorNombre    string `json:"compradorNombre"`
	CompradorApellidos string `json:"compradorApellidos"`
	VendedorDNI        string `json:"vendedorDni"`
	VendedorNombre     string `json:"vendedorNombre"`
	VendedorApellidos  string `json:"vendedorApellidos"`
	EmpleadoID         string `json:"empleadoId"`
	FechaAdquisicion   string `json:"fechaAdquisicion"`
	ArchivoNombre      string `json:"archivoNombre"`
	CreatedBy          string `json:"createdBy"`
}

const idCounterKey = "ID_COUNTER"

func (rt *RealEstateTransaction) getNextID(ctx contractapi.TransactionContextInterface) (int, error) {
	counterBytes, err := ctx.GetStub().GetState(idCounterKey)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo el contador de ID: %s", err)
	}

	var idCounter int
	if counterBytes == nil {
		idCounter = 1
	} else {
		idCounter, _ = strconv.Atoi(string(counterBytes))
		idCounter++
	}

	err = ctx.GetStub().PutState(idCounterKey, []byte(strconv.Itoa(idCounter)))
	if err != nil {
		return 0, fmt.Errorf("error actualizando el contador de ID: %s", err)
	}

	return idCounter, nil
}

func (rt *RealEstateTransaction) CreateTransaction(ctx contractapi.TransactionContextInterface, partidaRegistral, compradorDNI, compradorNombre, compradorApellidos, vendedorDNI, vendedorNombre, vendedorApellidos, empleadoID, archivoNombre string) error {
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("error obteniendo la marca de tiempo de la transacción: %s", err)
	}

	transactionID, err := rt.getNextID(ctx)
	if err != nil {
		return fmt.Errorf("error obteniendo el próximo ID: %s", err)
	}

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error obteniendo el ID del cliente: %s", err)
	}

	transaction := Transaction{
		ID:                 transactionID,
		PartidaRegistral:   partidaRegistral,
		CompradorDNI:       compradorDNI,
		CompradorNombre:    compradorNombre,
		CompradorApellidos: compradorApellidos,
		VendedorDNI:        vendedorDNI,
		VendedorNombre:     vendedorNombre,
		VendedorApellidos:  vendedorApellidos,
		EmpleadoID:         empleadoID,
		FechaAdquisicion:   time.Unix(timestamp.Seconds, 0).Format("2006-01-02"),
		ArchivoNombre:      archivoNombre,
		CreatedBy:          clientID,
	}

	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("error serializando la transacción: %s", err)
	}

	transactionKey := "TRANSACTION" + strconv.Itoa(transaction.ID)
	return ctx.GetStub().PutState(transactionKey, transactionJSON)
}

func (rt *RealEstateTransaction) GetTransaction(ctx contractapi.TransactionContextInterface, id int) (*Transaction, error) {
	transactionKey := "TRANSACTION" + strconv.Itoa(id)
	transactionBytes, err := ctx.GetStub().GetState(transactionKey)
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el estado para el ID %d: %s", id, err)
	}
	if transactionBytes == nil {
		return nil, fmt.Errorf("no se encontró la transacción con el ID %d", id)
	}

	var transaction Transaction
	err = json.Unmarshal(transactionBytes, &transaction)
	if err != nil {
		return nil, fmt.Errorf("error deserializando la transacción: %s", err)
	}

	return &transaction, nil
}

func (rt *RealEstateTransaction) DeleteTransaction(ctx contractapi.TransactionContextInterface, id int) error {
	transactionKey := "TRANSACTION" + strconv.Itoa(id)
	transaction, err := rt.GetTransaction(ctx, id)
	if err != nil {
		return err
	}

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error obteniendo el ID del cliente: %s", err)
	}

	if transaction.CreatedBy != clientID {
		return fmt.Errorf("solo el creador puede eliminar esta transacción")
	}

	return ctx.GetStub().DelState(transactionKey)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(RealEstateTransaction))
	if err != nil {
		fmt.Printf("Error creando el chaincode RealEstateTransaction: %s", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error iniciando el chaincode RealEstateTransaction: %s", err)
	}
}
