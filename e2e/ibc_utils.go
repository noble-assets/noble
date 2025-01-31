// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	chantypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
)

// ICATestSuite is used to test interchain accounts. The IcaAddress field is initially an empty string,
// after initializing the ICA account you must update the field with the corresponding value if you need to send ICA txs in later test logic.
type ICATestSuite struct {
	Host                   *cosmos.CosmosChain
	Controller             *cosmos.CosmosChain
	Relayer                ibc.Relayer
	Rep                    *testreporter.RelayerExecReporter
	OwnerAddress           string
	IcaAddress             string
	InitBal                math.Int
	HostConnectionID       string
	ControllerConnectionID string

	Encoding string
	Msgs     []sdk.Msg
	TxMemo   string
}

// RegisterICAAccount attempts to register a new interchain account on a host chain via a controller chain.
func RegisterICAAccount(ctx context.Context, icaTs *ICATestSuite) (string, error) {
	height, err := icaTs.Controller.Height(ctx)
	if err != nil {
		return "", err
	}

	version, err := json.Marshal(map[string]any{
		"version":                  icatypes.Version,
		"controller_connection_id": icaTs.ControllerConnectionID,
		"host_connection_id":       icaTs.HostConnectionID,
		"address":                  "",
		"encoding":                 icaTs.Encoding,
		"tx_type":                  icatypes.TxTypeSDKMultiMsg,
	})
	if err != nil {
		return "", err
	}

	cmd := []string{
		"interchain-accounts", "controller", "register",
		icaTs.ControllerConnectionID,
		"--ordering", "ORDER_ORDERED",
		"--version", string(version),
	}

	_, err = icaTs.Controller.Validators[0].ExecTx(ctx, icaTs.OwnerAddress, cmd...)
	if err != nil {
		return "", err
	}

	// It takes a few blocks for the ICA channel handshake to be completed by the relayer.
	// Query for the MsgChannelOpenConfirm on the host chain so we know when the ICA has been initialized.
	channelFound := func(found *chantypes.MsgChannelOpenConfirm) bool {
		return found.PortId == icatypes.HostPortID
	}

	_, err = cosmos.PollForMessage(ctx, icaTs.Host, NobleEncoding().InterfaceRegistry, height, height+15, channelFound)
	if err != nil {
		return "", fmt.Errorf("failed to poll for channel open confirmation: %w", err)
	}

	icaAddr, err := ICAAddress(ctx, icaTs.Controller, icaTs.OwnerAddress, icaTs.ControllerConnectionID)
	if err != nil {
		return "", err
	}

	err = icaTs.Host.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: icaAddr,
		Denom:   icaTs.Host.Config().Denom,
		Amount:  icaTs.InitBal,
	})
	if err != nil {
		return "", err
	}

	return icaAddr, nil
}

// ICAAddressResponse represents the response from querying an interchain account via the controller module.
type ICAAddressResponse struct {
	Address string `json:"address"`
}

// ICAAddress attempts to query the address of a registered interchain account on a controller chain for a given
// address and connection ID.
func ICAAddress(
	ctx context.Context,
	controller *cosmos.CosmosChain,
	address string,
	connectionID string,
) (string, error) {
	cmd := []string{
		"interchain-accounts", "controller", "interchain-account",
		address, connectionID,
	}

	stdout, _, err := controller.Validators[0].ExecQuery(ctx, cmd...)
	if err != nil {
		return "", err
	}

	var result ICAAddressResponse
	err = json.Unmarshal(stdout, &result)
	if err != nil {
		return "", err
	}

	if result.Address == "" {
		return "", fmt.Errorf("ICA address not found for address(%s), connection(%s)", address, connectionID)
	}

	return result.Address, nil
}

// SendICATx attempts to serialize a slice of sdk.Msg and generate interchain account packet data, using the
// specified encoding, before attempting to send an ICA tx via the controller module.
func SendICATx(ctx context.Context, icaTs *ICATestSuite) error {
	cdc := codec.NewProtoCodec(icaTs.Controller.GetCodec().InterfaceRegistry())

	dataBz, err := icatypes.SerializeCosmosTx(cdc, icaTs.Msgs, icaTs.Encoding)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: dataBz,
		Memo: icaTs.TxMemo,
	}

	if err := packetData.ValidateBasic(); err != nil {
		return err
	}

	packetDataBz, err := cdc.MarshalJSON(&packetData)
	if err != nil {
		return err
	}

	node := icaTs.Controller.Validators[0]

	relPath := "packet_msg.json"
	err = node.WriteFile(ctx, packetDataBz, relPath)
	if err != nil {
		return err
	}

	packetMsgPath := path.Join(node.HomeDir(), relPath)

	height, err := icaTs.Controller.Height(ctx)
	if err != nil {
		return err
	}

	cmd := []string{
		"interchain-accounts", "controller", "send-tx",
		icaTs.ControllerConnectionID,
		packetMsgPath,
	}

	_, err = node.ExecTx(ctx, icaTs.OwnerAddress, cmd...)
	if err != nil {
		return err
	}

	_, err = cosmos.PollForMessage[*chantypes.MsgAcknowledgement](ctx, icaTs.Controller, NobleEncoding().InterfaceRegistry, height, height+20, nil)
	if err != nil {
		return fmt.Errorf("failed to poll for acknowledgement: %w", err)
	}

	return nil
}
