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

package accounts

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"noble.xyz/x/accounts/types"
	"noble.xyz/x/accounts/types/ibc"
	ibclegacy "noble.xyz/x/accounts/types/ibc/legacy"
)

func MigrateLegacyAccounts(ctx context.Context, logger log.Logger, accountKeeper types.AccountKeeper) {
	migratedIbcCount := 0

	accountKeeper.IterateAccounts(ctx, func(account sdk.AccountI) bool {
		if forwardingAccount, ok := account.(*ibclegacy.ForwardingAccount); ok {
			pubKey := types.PubKey{Key: forwardingAccount.GetAddress()}
			// TODO(@john): Figure out how to gracefully handle this error!
			rawPubKey, _ := codectypes.NewAnyWithValue(&pubKey)

			baseAccount := forwardingAccount.BaseAccount
			baseAccount.PubKey = rawPubKey

			attributes := &ibc.Attributes{
				Channel:   forwardingAccount.Channel,
				Recipient: forwardingAccount.Recipient,
				Fallback:  forwardingAccount.Fallback,
			}
			// TODO(@john): Figure out how to gracefully handle this error!
			rawAttributes, _ := codectypes.NewAnyWithValue(attributes)

			newAccount := types.NewAccount(forwardingAccount.BaseAccount, rawAttributes)

			accountKeeper.SetAccount(ctx, newAccount)
			migratedIbcCount++

			return false
		}

		return false
	})

	logger.Info(fmt.Sprintf("migrated %d forwarding accounts", migratedIbcCount))
}
