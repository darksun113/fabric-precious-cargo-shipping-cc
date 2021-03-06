// fabric-precious-cargo-shipping-cc is a sample chaincode for Hyperledger Fabric
// Copyright (C) 2019 @aschmidt75
package main

import (
	"encoding/json"
	"errors"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"time"
)

// creates a new shipment structure from given Ids of shipper and Participants.
type submitShipmentArg struct {
	Shipper     string `json:"by"`
	From        string `json:"from"`
	To          string `json:"to"`
	SubmittedAt string `json:"submittedAt"`
}

// Returns ID of shipment
type submitShipmentResult struct {
	ID string `json:"id"`
}

type submitShipmentInvocation struct {
	// input argument
	arg submitShipmentArg

	// intermediates
	shipperKey, fromKey, toKey string
	submittedAtParsed          time.Time

	// result
	res submitShipmentResult
}

func (inv *submitShipmentInvocation) checkParseArguments(stub shim.ChaincodeStubInterface) error {
	logger.Println("enter submitShipmentInvocation.checkParseArguments")

	_, args := stub.GetFunctionAndParameters()

	if len(args) != 1 {
		return errors.New("expecting JSON input as first param")
	}

	inv.arg = submitShipmentArg{}
	err := json.Unmarshal([]byte(args[0]), &inv.arg)
	if err != nil {
		logger.Printf("Error unmarshaling JSON: %s", err)
		return errors.New("Invalid JSON")
	}

	// check IDs
	var k string
	k, _, err = shipmentCoRegistry().get(stub, inv.arg.Shipper)
	if err != nil {
		logger.Println(err)
		return errors.New("invalid shipper argument: Not found")
	}
	inv.shipperKey = k

	k, _, err = individualParticipantRegistry().get(stub, inv.arg.From)
	if err != nil {
		return errors.New("invalid from argument: Not found")
	}
	inv.fromKey = k

	k, _, err = individualParticipantRegistry().get(stub, inv.arg.To)
	if err != nil {
		return errors.New("invalid to argument: Not found")
	}
	inv.toKey = k

	// parse and check time
	inv.submittedAtParsed, err = time.Parse(time.RFC3339, inv.arg.SubmittedAt)
	if err != nil {
		return errors.New("invalid submittedAt argument: Not parseable, please provide in RFC3339, e.g. 2006-01-02T15:04:05Z")
	}
	logger.Printf("Parsed submittedAt=%s\n", inv.submittedAtParsed)

	// check against now, e.g. diff must be <1h or the like
	difference := time.Now().Sub(inv.submittedAtParsed)
	logger.Printf("Diff to now is=%s\n", difference)

	return nil
}

func (inv *submitShipmentInvocation) process(stub shim.ChaincodeStubInterface) error {
	logger.Println("enter submitShipmentInvocation.process")
	logger.Printf("arg=%#v\n", inv.arg)

	key, err := shipmentRegistry().create(stub, Shipment{
		ShipperID:   inv.arg.Shipper,
		FromID:      inv.arg.From,
		ToID:        inv.arg.To,
		Status:      "submitted",
		SubmittedAt: inv.submittedAtParsed,
	})
	if err != nil {
		return errors.New("internal error writing world state")
	}
	inv.res = submitShipmentResult{
		ID: key,
	}

	return nil
}

func (inv *submitShipmentInvocation) getResponse(stub shim.ChaincodeStubInterface) interface{} {
	return inv.res
}
