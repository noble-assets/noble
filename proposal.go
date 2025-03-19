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
	"context"
	"fmt"
	"slices"
	"time"

	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"

	"connectrpc.com/connect"
	jester "jester.noble.xyz/api"

	wormholekeeper "github.com/noble-assets/wormhole/keeper"
	wormholetypes "github.com/noble-assets/wormhole/types"
	vaautils "github.com/wormhole-foundation/wormhole/sdk/vaa"

	dollarkeeper "dollar.noble.xyz/keeper"
	dollarportaltypes "dollar.noble.xyz/types/portal"
)

// jesterIndex is the index of the injected Jester response in a block.
const jesterIndex = 0

type ProposalHandler struct {
	txConfig client.TxConfig

	jesterClient   jester.QueryServiceClient
	wormholeServer wormholetypes.QueryServer
	dollarKeeper   *dollarkeeper.Keeper

	defaultPrepareProposalHandler sdk.PrepareProposalHandler
	defaultPreBlocker             sdk.PreBlocker
}

func NewProposalHandler(
	app *baseapp.BaseApp,
	mempool mempool.Mempool,
	preBlocker sdk.PreBlocker,
	txConfig client.TxConfig,
	jesterClient jester.QueryServiceClient,
	dollarKeeper *dollarkeeper.Keeper,
	wormholeKeeper *wormholekeeper.Keeper,
) *ProposalHandler {
	defaultHandler := baseapp.NewDefaultProposalHandler(mempool, app)

	return &ProposalHandler{
		txConfig: txConfig,

		jesterClient:   jesterClient,
		wormholeServer: wormholekeeper.NewQueryServer(wormholeKeeper),
		dollarKeeper:   dollarKeeper,

		defaultPrepareProposalHandler: defaultHandler.PrepareProposalHandler(),
		defaultPreBlocker:             preBlocker,
	}
}

// PrepareProposal is the logic called by the current block proposer to prepare
// a block proposal. Noble modifies this by making a request to our sidecar
// service, Jester, to check if there are any outstanding $USDN transfers that
// need to be relayed to Noble. These transfers (in the form of Wormhole VAAs)
// are injected as the first transaction of the block, and are later processed
// by the PreBlocker handler.
func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		logger := ctx.Logger()

		res, err := h.defaultPrepareProposalHandler(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "default PrepareProposal handler failed")
		}

		ctxWithTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		request := connect.NewRequest(&jester.GetVoteExtensionRequest{})
		jesterRes, err := h.jesterClient.GetVoteExtension(ctxWithTimeout, request)
		if err != nil {
			logger.Error("failed to query jester", "err", err)
		}

		if jesterRes != nil && jesterRes.Msg != nil && jesterRes.Msg.Dollar != nil && len(jesterRes.Msg.Dollar.Vaas) > 0 {
			var nonExecutedVAAs []sdk.Msg

			for _, raw := range jesterRes.Msg.Dollar.Vaas {
				vaa, err := vaautils.Unmarshal(raw)
				if err != nil {
					logger.Warn("failed to unmarshal transfer from jester", "err", err)
					continue
				}

				wormholeRes, _ := h.wormholeServer.ExecutedVAA(ctx, &wormholetypes.QueryExecutedVAA{
					Input: vaa.SigningDigest().String(),
				})

				if wormholeRes != nil && !wormholeRes.Executed {
					nonExecutedVAAs = append(nonExecutedVAAs, &dollarportaltypes.MsgDeliverInjection{
						Vaa: raw,
					})
				} else {
					logger.Warn("skipped already executed transfer from jester", "identifier", vaa.MessageID())
				}
			}

			if len(nonExecutedVAAs) > 0 {
				builder := h.txConfig.NewTxBuilder()

				err := builder.SetMsgs(nonExecutedVAAs...)
				if err != nil {
					return nil, errors.Wrap(err, "failed to set messages of injected jester tx")
				}

				tx := builder.GetTx()

				bz, err := h.txConfig.TxEncoder()(tx)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal injected jester tx")
				}
				res.Txs = slices.Insert(res.Txs, jesterIndex, bz)

				logger.Info(fmt.Sprintf("injected %d pending transfers from jester", len(nonExecutedVAAs)))
			}
		}

		return &abci.ResponsePrepareProposal{Txs: res.Txs}, nil
	}
}

// PreBlocker processes all injected $USDN transfers from Jester.
func (h *ProposalHandler) PreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		res, err := h.defaultPreBlocker(ctx, req)
		if err != nil {
			return nil, errors.Wrap(err, "default PreBlocker failed")
		}

		if len(req.Txs) == 0 {
			return res, nil
		}

		tx := req.Txs[jesterIndex]
		h.handleJesterTx(ctx, tx)

		return res, nil
	}
}

// handleJesterTx is a utility that processes messages from Jester.
func (h *ProposalHandler) handleJesterTx(ctx sdk.Context, bytes []byte) {
	logger := ctx.Logger()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("recovered panic when handling transfers from jester", "err", r)
		}
	}()

	tx, err := h.txConfig.TxDecoder()(bytes)
	if err != nil {
		logger.Error("failed to unmarshal injected jester tx", "err", err)
		return
	}

	var count int
	for _, raw := range tx.GetMsgs() {
		msg, ok := raw.(*dollarportaltypes.MsgDeliverInjection)
		// If the first message is not a MsgDeliverInjection, no VAAs were injected.
		if !ok {
			break
		}

		vaa, err := vaautils.Unmarshal(msg.Vaa)
		if err != nil {
			logger.Error("failed to unmarshal transfer from jester", "err", err)
			continue
		}

		cachedCtx, writeCache := ctx.CacheContext()
		if err := h.dollarKeeper.Deliver(cachedCtx, msg.Vaa); err != nil {
			logger.Error("failed to process transfer from jester", "identifier", vaa.MessageID(), "err", err)
		} else {
			writeCache()
			count++
		}

	}
	if count > 0 {
		logger.Info(fmt.Sprintf("processed %d transfers from jester", count))
	}
}
