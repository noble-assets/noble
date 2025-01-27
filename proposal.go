package noble

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"connectrpc.com/connect"
	"cosmossdk.io/log"
	dollarkeeper "dollar.noble.xyz/keeper"
	dollarportaltypes "dollar.noble.xyz/types/portal"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	jester "jester.noble.xyz/api"
)

const (
	// maximum amount of blocks a validator is allowed to submit without a jester response
	maxNoJesterBlocks = uint16(1000)
)

type ProposalHandler struct {
	logger       log.Logger
	bApp         *baseapp.BaseApp
	jester       jester.QueryServiceClient
	dollarKeeper *dollarkeeper.Keeper

	noJesterBlocks uint16

	defaultPrepareProposalHandler sdk.PrepareProposalHandler
	defaultProcessProposalHandler sdk.ProcessProposalHandler
}

func NewProposalHandler(
	logger log.Logger,
	bApp *baseapp.BaseApp,
	mp mempool.Mempool,
	jester jester.QueryServiceClient,
	dollarKeeper *dollarkeeper.Keeper,
) *ProposalHandler {
	defaultHandler := baseapp.NewDefaultProposalHandler(mp, bApp)
	return &ProposalHandler{
		logger:       logger,
		bApp:         bApp,
		jester:       jester,
		dollarKeeper: dollarKeeper,

		defaultPrepareProposalHandler: defaultHandler.PrepareProposalHandler(),
		defaultProcessProposalHandler: defaultHandler.ProcessProposalHandler(),
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		log := ctx.Logger()
		fmt.Println("IN PREPARE PROPOSAL!!", "HEIGHT", req.Height)

		proposalTxs := req.Txs

		// call default handler first for basic validation
		res, err := h.defaultPrepareProposalHandler(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed in default PrepareProposal handler: %w", err)
		}

		request := connect.NewRequest(&jester.GetVoteExtensionRequest{})
		jRes, err := h.jester.GetVoteExtension(context.Background(), request)
		// TODO: consider better handling of "noJesterBlocks" logic
		if err != nil {
			h.noJesterBlocks++

			log.Error(`failed to get vote extention from Jester. Your node will panic if "MAX-noJesterBlocks" is reached!!!`,
				"current-noJesterBlocks", h.noJesterBlocks, "MAX-noJesterBlocks", maxNoJesterBlocks,
				"err", err,
			)

			if h.noJesterBlocks >= maxNoJesterBlocks {
				panic("too many consecutive blocks porposed without jester response")
			}

			return res, nil
		}

		// if jester response is received, reset noJesterBlocks counter
		h.noJesterBlocks = 0

		// TODO: use wormhole keeper to check if VAA has already been submitted and adjust list.
		// Then we can remove duplicate code in Jester.

		bz, err := json.Marshal(jRes.Msg)
		if err != nil {
			return res, fmt.Errorf("failed to marshal jester response: %w", err)
		}

		// inject into block
		proposalTxs = slices.Insert(proposalTxs, 0, bz)

		ctx.Logger().Info("received vote extension", "num_vaas", len(jRes.Msg.Dollar.Vaas))

		return &abci.ResponsePrepareProposal{Txs: proposalTxs}, nil
	}
}

func (h *ProposalHandler) NewProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		fmt.Println("IN PROCESS PROPOSAL!!", "HEIGHT", req.Height)

		resAccept := &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
		// resReject := &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}

		if len(req.Txs) == 0 {
			fmt.Println("NO TXS!") // TODO: remove
			return resAccept, nil
		}

		// TODO: Do checks?
		// TODO: defaultProcessProposalHandler?
		// TODO: test how normal non-injected tx's are handled

		return resAccept, nil
	}
}

func (h *ProposalHandler) NewPreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		fmt.Println("IN PRE BLOCKER!!!!", "HEIGHT", req.Height)

		res := &sdk.ResponsePreBlock{}
		if len(req.Txs) == 0 {
			return res, nil
		}

		// TODO: is this ok if there aren't any VAA's to inject but there are other normal tx's?
		var jesterResponse jester.GetVoteExtensionResponse
		if err := json.Unmarshal(req.Txs[0], &jesterResponse); err != nil {
			ctx.Logger().Error("failed to decode injected dollar vote extension", "err", err)
			return nil, err
		}

		var successfullMsg int
		server := dollarkeeper.NewPortalMsgServer(h.dollarKeeper)
		for _, vaa := range jesterResponse.Dollar.Vaas {
			cachedCtx, writeCache := ctx.CacheContext()
			_, err := server.Deliver(cachedCtx, &dollarportaltypes.MsgDeliver{
				Vaa: vaa,
			})

			if err == nil {
				writeCache()
				successfullMsg++
			} else {
				ctx.Logger().Error("failed to submit VAA", "err", err)
			}
		}

		if successfullMsg > 0 {
			ctx.Logger().Info("successfully submitted VAAs", "num_vaas", successfullMsg)
		}

		return res, nil
	}
}
