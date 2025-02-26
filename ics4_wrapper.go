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

package noble

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

var _ porttypes.ICS4Wrapper = &NobleICS4Wrapper{}

// NobleICS4Wrapper implements the ICS4Wrapper interface. It implements custom logic in SendPacket in order
// to check all outgoing IBC transfers so that $USDN cannot be sent to another chain.
type NobleICS4Wrapper struct {
	ics4Wrapper  porttypes.ICS4Wrapper
	dollarKeeper ExpectedDollarKeeper
}

// ExpectedDollarKeeper defines the interface expected by NobleICS4Wrapper for the Noble Dollar module.
type ExpectedDollarKeeper interface {
	GetDenom() string
}

// NewNobleICS4Wrapper returns a new instance of NobleICS4Wrapper.
func NewNobleICS4Wrapper(app porttypes.ICS4Wrapper, dollarKeeper ExpectedDollarKeeper) porttypes.ICS4Wrapper {
	return NobleICS4Wrapper{
		ics4Wrapper:  app,
		dollarKeeper: dollarKeeper,
	}
}

// SendPacket attempts to unmarshal the provided packet data as the ICS-20
// FungibleTokenPacketData type. If the packet is a valid ICS-20 transfer, then
// a check is performed on the denom to ensure that $USDN cannot be transferred
// out of Noble via IBC.
func (w NobleICS4Wrapper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (sequence uint64, err error) {
	var packetData transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(data, &packetData); err != nil {
		return w.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	denom := w.dollarKeeper.GetDenom()
	if packetData.Denom == denom {
		return 0, fmt.Errorf("ibc transfers of %s are currently disabled", denom)
	}

	return w.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement implements the ICS4Wrapper interface.
func (w NobleICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return w.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// GetAppVersion implements the ICS4Wrapper interface.
func (w NobleICS4Wrapper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return w.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
