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

package jester

import (
	"fmt"

	dollarkeeper "dollar.noble.xyz/v2/keeper"
	dollarportaltypes "dollar.noble.xyz/v2/types/portal"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vaautils "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func PreBlockerHandler(txConfig client.TxConfig, dollarKeeper *dollarkeeper.Keeper) sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		injection := parseInjection(req.Txs, txConfig.TxDecoder())
		if len(injection) != 0 {
			var count int
			for _, msg := range injection {
				vaa, err := vaautils.Unmarshal(msg.Vaa)
				if err != nil {
					ctx.Logger().Error("failed to unmarshal transfer from jester", "err", err)
					continue
				}

				cachedCtx, writeCache := ctx.CacheContext()
				if err := dollarKeeper.Deliver(cachedCtx, msg.Vaa); err != nil {
					ctx.Logger().Error("failed to process transfer from jester", "identifier", vaa.MessageID(), "err", err)
				} else {
					writeCache()
					count++
				}
			}
			if count > 0 {
				ctx.Logger().Info(fmt.Sprintf("processed %d transfers from jester", count))
			}
		}

		return &sdk.ResponsePreBlock{ConsensusParamsChanged: false}, nil
	}
}

func parseInjection(txs [][]byte, txDecoder sdk.TxDecoder) []*dollarportaltypes.MsgDeliverInjection {
	// Because both Nova and Jester optionally inject transactions, we have to
	// handle all three different cases of injections.
	limit := len(txs)
	maxRange := 2
	if limit > maxRange {
		limit = maxRange
	}

	for _, tx := range txs[:limit] {
		if inj := parseInjectionFromTx(tx, txDecoder); len(inj) != 0 {
			return inj
		}
	}

	return nil
}

func parseInjectionFromTx(bz []byte, txDecoder sdk.TxDecoder) []*dollarportaltypes.MsgDeliverInjection {
	tx, err := txDecoder(bz)
	if err != nil {
		return nil
	}

	var res []*dollarportaltypes.MsgDeliverInjection
	for _, raw := range tx.GetMsgs() {
		msg, ok := raw.(*dollarportaltypes.MsgDeliverInjection)
		// If the first message is not a MsgDeliverInjection, no VAAs were injected.
		if !ok {
			break
		}

		res = append(res, msg)
	}

	return res
}
