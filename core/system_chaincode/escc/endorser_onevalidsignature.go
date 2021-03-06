/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package escc

import (
	"errors"

	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("escc")

// EndorserOneValidSignature implements the default endorsement policy, which is to
// sign the proposal hash and the read-write set
type EndorserOneValidSignature struct {
}

// Init is called once when the chaincode started the first time
func (e *EndorserOneValidSignature) Init(stub shim.ChaincodeStubInterface) ([]byte, error) {
	// best practice to do nothing (or very little) in Init
	return nil, nil
}

// Invoke is called to endorse the specified Proposal
// For now, we sign the input and return the endorsed result. Later we can expand
// the chaincode to provide more sophisticate policy processing such as enabling
// policy specification to be coded as a transaction of the chaincode and Client
// could select which policy to use for endorsement using parameter
// @return a marshalled proposal response
// Note that Peer calls this function with 4 mandatory arguments (and 2 optional ones):
// args[0] - function name (not used now)
// args[1] - serialized Header object
// args[2] - serialized ChaincodeProposalPayload object
// args[3] - binary blob of simulation results
// args[4] - serialized events (optional)
// args[5] - payloadVisibility (optional)
//
// NOTE: this chaincode is meant to sign another chaincode's simulation
// results. It should not manipulate state as any state change will be
// silently discarded: the only state changes that will be persisted if
// this endorsement is successful is what we are about to sign, which by
// definition can't be a state change of our own.
func (e *EndorserOneValidSignature) Invoke(stub shim.ChaincodeStubInterface) ([]byte, error) {
	args := stub.GetArgs()
	if len(args) < 4 {
		return nil, fmt.Errorf("Incorrect number of arguments (expected a minimum of 4, provided %d)", len(args))
	} else if len(args) > 6 {
		return nil, fmt.Errorf("Incorrect number of arguments (expected a maximum of 6, provided %d)", len(args))
	}

	logger.Infof("ESCC starts: %d args", len(args))

	// handle the header
	var hdr []byte
	if args[1] == nil {
		return nil, errors.New("serialized Header object is null")
	} else {
		hdr = args[1]
	}

	// handle the proposal payload
	var payl []byte
	if args[2] == nil {
		return nil, errors.New("serialized ChaincodeProposalPayload object is null")
	} else {
		payl = args[2]
	}

	// handle simulation results
	var results []byte
	if args[3] == nil {
		return nil, errors.New("simulation results are null")
	} else {
		results = args[3]
	}

	// Handle serialized events if they have been provided
	// they might be nil in case there's no events but there
	// is a visibility field specified as the next arg
	events := []byte("")
	if len(args) > 4 && args[4] != nil {
		events = args[4]
	}

	// Handle payload visibility (it's an optional argument)
	visibility := []byte("") // TODO: when visibility is properly defined, replace with the default
	if len(args) > 5 {
		if args[5] == nil {
			return nil, errors.New("serialized events are null")
		} else {
			visibility = args[5]
		}
	}

	// obtain the proposal hash given proposal header, payload and the requested visibility
	pHashBytes, err := utils.GetProposalHash(hdr, payl, visibility)
	if err != nil {
		return nil, fmt.Errorf("Could not compute proposal hash: err %s", err)
	}

	// TODO: obtain current epoch
	epoch := []byte("current_epoch")
	logger.Infof("using epoch %s", string(epoch))

	// get the bytes of the proposal response payload - we need to sign them
	prpBytes, err := utils.GetBytesProposalResponsePayload(pHashBytes, epoch, results, events)
	if err != nil {
		return nil, errors.New("Failure while unmarshalling the ProposalResponsePayload")
	}

	// TODO: obtain the signing key for this endorser - what API should be used?
	endorser := []byte("here_goes_the_endorsers_key")

	// TODO: sign prpBytes with this endorser's key - use msp interfaces and providers
	signature := []byte("here_goes_the_signature_of_prpBytes_under_the_endorsers_key")

	// marshall the proposal response so that we return its bytes
	prBytes, err := utils.GetBytesProposalResponse(prpBytes, &protos.Endorsement{Signature: signature, Endorser: endorser})
	if err != nil {
		return nil, fmt.Errorf("Could not marshall ProposalResponse: err %s", err)
	}

	logger.Infof("ESCC exits successfully")
	return prBytes, nil
}

// Query is here to satisfy the Chaincode interface. We don't need it for this system chaincode
func (e *EndorserOneValidSignature) Query(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return nil, nil
}
