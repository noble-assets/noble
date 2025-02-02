package noble

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"connectrpc.com/connect"
	"cosmossdk.io/log"
	dollarkeeper "dollar.noble.xyz/keeper"
	dollarportaltypes "dollar.noble.xyz/types/portal"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	wormholekeeper "github.com/noble-assets/wormhole/keeper"
	wormholetypes "github.com/noble-assets/wormhole/types"
	vaautils "github.com/wormhole-foundation/wormhole/sdk/vaa"

	jester "jester.noble.xyz/api"
)

const (
	// index to inject jester response into block
	injectIndex = 0
)

type ProposalHandler struct {
	logger         log.Logger
	bApp           *baseapp.BaseApp
	jesterCli      jester.QueryServiceClient
	dollarKeeper   *dollarkeeper.Keeper
	wormholeKeeper *wormholekeeper.Keeper

	defaultPrepareProposalHandler sdk.PrepareProposalHandler
	defaultProcessProposalHandler sdk.ProcessProposalHandler
}

func NewProposalHandler(
	logger log.Logger,
	bApp *baseapp.BaseApp,
	mp mempool.Mempool,
	jesterCli jester.QueryServiceClient,
	dollarKeeper *dollarkeeper.Keeper,
	wormholeKeeper *wormholekeeper.Keeper,
) *ProposalHandler {
	defaultHandler := baseapp.NewDefaultProposalHandler(mp, bApp)
	return &ProposalHandler{
		logger:         logger,
		bApp:           bApp,
		jesterCli:      jesterCli,
		dollarKeeper:   dollarKeeper,
		wormholeKeeper: wormholeKeeper,

		defaultPrepareProposalHandler: defaultHandler.PrepareProposalHandler(),
		defaultProcessProposalHandler: defaultHandler.ProcessProposalHandler(),
	}
}

// PrepareProposal is called only by the proposing validator and prepares a proposal
// for the next block.
// It calls Jester to check if there are any outstanding VAAs. These VAAs are injected
// as bytes into the first transaction of the block and will later be handled by the PreBlocker.
func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		log := ctx.Logger()

		// Call default handler first for basic validation
		res, err := h.defaultPrepareProposalHandler(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed in default PrepareProposal handler: %w", err)
		}

		// Query Jester for VAA's
		request := connect.NewRequest(&jester.GetVoteExtensionRequest{})
		ctxTO, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		jRes, err := h.jesterCli.GetVoteExtension(ctxTO, request)
		if err != nil {
			log.Error(`failed to get vote extension from Jester. Ensure Jester is configured and running!`,
				"err", err,
			)
		}

		// If there are no transactions in the block, we do not require a Jester response.
		// In the PreBlocker, if there are transactions, we assume the first one is a Jester response
		// and handle it accordingly. If there are no transactions, this handling is not needed.
		// This allows us to retain the capability of having empty blocks which will help with chain bloat.
		requireJesterInj := len(res.Txs) > 0

		switch {
		// Jester response failed and we do not need to inject an empty Jester response
		case !requireJesterInj && jRes == nil:
			// no op
		// Jester response failed OR Jester responded with zero VAA's but we need to inject an empty Jester response
		case requireJesterInj && jRes == nil || requireJesterInj && jRes.Msg.Dollar.Vaas == nil:
			emptyJesterRes, err := json.Marshal(jester.GetVoteExtensionResponse{})
			if err != nil {
				log.Error("failed to marshal empty jester response", "err", err)
			}
			res.Txs = slices.Insert(res.Txs, injectIndex, emptyJesterRes)
		// Jester responded with VAA's that need to be injected
		case jRes.Msg.Dollar.Vaas != nil:
			var nonExecutedVaas [][]byte

			// Check if the VAA's have already been executed.
			wormholeSever := wormholekeeper.NewQueryServer(h.wormholeKeeper)
			for _, raw := range jRes.Msg.Dollar.Vaas {
				vaa, err := vaautils.Unmarshal(raw)
				if err != nil {
					log.Warn("failed to unmarshal vaa from jester", "err", err)
				}

				r, _ := wormholeSever.ExecutedVAA(ctx, &wormholetypes.QueryExecutedVAA{
					Input: vaa.SigningDigest().String(),
				})

				if r != nil && !r.Executed {
					nonExecutedVaas = append(nonExecutedVaas, raw)
				} else {
					// TODO: Keeping this log in for testing purposes only. This should be removed for final release.
					log.Info("received already executed vaa from jester", "identifier", vaa.MessageID())
				}
			}

			if len(nonExecutedVaas) > 0 {
				jRes.Msg.Dollar.Vaas = nonExecutedVaas

				bz, err := json.Marshal(jRes.Msg)
				if err != nil {
					return res, fmt.Errorf("failed to marshal jester response: %w", err)
				}

				// inject VAA bytes into block. These will be handled in the PreBlocker
				res.Txs = slices.Insert(res.Txs, injectIndex, bz)
				ctx.Logger().Info("received vote extension", "num_vaas", len(jRes.Msg.Dollar.Vaas))
			}
		}

		return &abci.ResponsePrepareProposal{Txs: res.Txs}, nil
	}
}

// ProcessProposal validates the proposed block along with the transactions. This is called by
// all validators except for the proposer.
// It returns a status if it was accepted or rejected.
// We currently do not validate the injected bytes from the proposalHandler as these bytes will be
// handled in the PreBlocker.
func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		resAccept := &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
		resReject := &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}

		if len(req.Txs) == 0 {
			return resAccept, nil
		}

		// Remove injected Jester bytes to perform basic validation on other transactions
		req.Txs = append(req.Txs[:injectIndex], req.Txs[injectIndex+1:]...)
		res, err := h.defaultProcessProposalHandler(ctx, req)
		if err != nil {
			return resReject, fmt.Errorf("failed in default ProcessProposal handler: %w", err)
		}
		if !res.IsAccepted() {
			h.logger.Error("the proposal is rejected by default ProcessProposal handler",
				"height", req.Height)
			return resReject, nil
		}

		// TODO: Consult:
		// Should we validate the injected bytes?
		// I don't think this is necessary since this is going through a message server.
		// It is my understanding that rejecting the proposal will reject all transactions in this block.

		return resAccept, nil
	}
}

// NewPreBlocker submits the VAA's using a message server tied to the dollar keeper.
// This is called by all validators.
func (h *ProposalHandler) NewPreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		// There are no transactions in this block, no need to parse and submit any VAA's
		res := &sdk.ResponsePreBlock{}
		if len(req.Txs) == 0 {
			return res, nil
		}

		var jesterResponse jester.GetVoteExtensionResponse
		if err := json.Unmarshal(req.Txs[0], &jesterResponse); err != nil {
			ctx.Logger().Error("failed to decode injected dollar vote extension", "err", err)
			return nil, err
		}

		if jesterResponse.Dollar != nil {
			var successfullMsgs int
			dollarServer := dollarkeeper.NewPortalMsgServer(h.dollarKeeper)
			for _, vaa := range jesterResponse.Dollar.Vaas {
				cachedCtx, writeCache := ctx.CacheContext()
				_, err := dollarServer.Deliver(cachedCtx, &dollarportaltypes.MsgDeliver{
					Vaa: vaa,
				})

				if err == nil {
					writeCache()
					successfullMsgs++
				} else {
					ctx.Logger().Error("failed to submit VAA", "err", err)
				}
			}

			if successfullMsgs > 0 {
				ctx.Logger().Info("successfully submitted VAAs", "num_vaas", successfullMsgs)
			}
		}

		return res, nil
	}
}
