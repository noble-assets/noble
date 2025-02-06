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
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/stretchr/testify/require"
)

var _ porttypes.ICS4Wrapper = (*MockICS4Wrapper)(nil)

type MockICS4Wrapper struct {
	t *testing.T
}

func (m MockICS4Wrapper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return 0, nil
}

func (m MockICS4Wrapper) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	m.t.Fatal("WriteAcknowledgement should not have been called")
	return nil
}

func (m MockICS4Wrapper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	m.t.Fatal("GetAppVersion should not have been called")
	return "", false
}

type MockDollarKeeper struct {
	denom string
}

func (m MockDollarKeeper) GetDenom() string {
	return m.denom
}

// TestSendPacket asserts that outgoing IBC transfers work as expected in cases
// where the denom is $USDN, as well as cases where the denom is not.
func TestSendPacket(t *testing.T) {
	denom := "uusdn"

	tc := []struct {
		name string
		data transfertypes.FungibleTokenPacketData
		fail bool
	}{
		{
			"Outgoing IBC transfer of USDN - should be blocked",
			transfertypes.NewFungibleTokenPacketData(denom, "1000000", "test", "test", "test"),
			true,
		},
		{
			"Outgoing IBC transfer of USDC - should not be blocked",
			transfertypes.NewFungibleTokenPacketData("uusdc", "1000000", "test", "test", "test"),
			false,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := MockICS4Wrapper{t}
			keeper := MockDollarKeeper{denom: denom}
			nobleWrapper := NewNobleICS4Wrapper(wrapper, keeper)

			data, err := transfertypes.ModuleCdc.MarshalJSON(&tt.data)
			require.NoError(t, err)

			ctx := sdk.Context{}
			timeout := uint64(0)

			_, err = nobleWrapper.SendPacket(ctx, nil, "transfer", "channel-0", clienttypes.Height{}, timeout, data)

			if tt.fail {
				require.Error(t, err)
				require.ErrorContains(t, err, fmt.Sprintf("ibc transfers of %s are currently disabled", denom))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
