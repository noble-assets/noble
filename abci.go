package noble

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"connectrpc.com/connect"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wormholekeeper "github.com/noble-assets/wormhole/keeper"
	wormholetypes "github.com/noble-assets/wormhole/types"
	jester "jester.noble.xyz/api"
)

// InjectedDollarVE is the vote extension data for the Noble Dollar that is injected as a fake transaction into a block.
type InjectedDollarVE struct {
	VAAs [][]byte
}

// client queryv1connect.QueryServiceClient
func NewExtendVoteHandler(client jester.QueryServiceClient) sdk.ExtendVoteHandler {
	// Returns custom functionality for the block proposer to execute.
	return func(ctx sdk.Context, req *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
		log := ctx.Logger()

		request := connect.NewRequest(&jester.GetVoteExtensionRequest{})
		res, err := client.GetVoteExtension(context.Background(), request)
		if err != nil {
			log.Error("failed to get vote extention from jester", "err", err)
		}

		bz, err := json.Marshal(res.Msg)
		if err != nil {
			log.Error("failed to marshal vote extention", "err", err)
		}

		return &abcitypes.ResponseExtendVote{
			VoteExtension: bz,
		}, err
	}
}

func NewVerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	// Returns custom functionality for other validators, not the current proposer, to execute.
	return func(ctx sdk.Context, req *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
		var voteExtension jester.GetVoteExtensionResponse

		err := json.Unmarshal(req.VoteExtension, &voteExtension)
		if err != nil {
			// NOTE: It is safe to return an error as the Cosmos SDK will capture all
			// errors, log them, and reject the proposal.
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)

		}

		ctx.Logger().Info("verifying vote extension", "num_vaas", len(voteExtension.Dollar.Vaas))

		return &abcitypes.ResponseVerifyVoteExtension{
			Status: abcitypes.ResponseVerifyVoteExtension_ACCEPT,
		}, nil
	}
}

func NewPrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
		if req.Height <= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			return &abcitypes.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		rawVoteExtension := []byte{}
		highestPowerVal := int64(math.MinInt64)

		for _, vote := range req.LocalLastCommit.Votes {
			if vote.Validator.Power > highestPowerVal {
				rawVoteExtension = vote.VoteExtension
				highestPowerVal = vote.Validator.Power
			}
		}

		var voteExtension jester.GetVoteExtensionResponse
		err := json.Unmarshal(rawVoteExtension, &voteExtension)
		if err != nil {
			return nil, err
		}

		bz, err := json.Marshal(InjectedDollarVE{
			VAAs: voteExtension.Dollar.Vaas,
		})
		if err != nil {
			return nil, err
		}
		req.Txs = append([][]byte{bz}, req.Txs...)

		ctx.Logger().Info("received vote extension", "num_vaas", len(voteExtension.Dollar.Vaas))

		return &abcitypes.ResponsePrepareProposal{Txs: req.Txs}, nil
	}
}

func NewPreBlocker(wormholeKeeper *wormholekeeper.Keeper) sdk.PreBlocker {
	return func(ctx sdk.Context, req *abcitypes.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		res := &sdk.ResponsePreBlock{}
		if len(req.Txs) == 0 {
			return res, nil
		}

		var dollarVoteExtension InjectedDollarVE
		if err := json.Unmarshal(req.Txs[0], &dollarVoteExtension); err != nil {
			ctx.Logger().Error("failed to decode injected dollar vote extension", "err", err)
			return nil, err
		}

		server := wormholekeeper.NewMsgServer(wormholeKeeper)
		for _, vaa := range dollarVoteExtension.VAAs {
			cachedCtx, writeCache := ctx.CacheContext()
			_, err := server.SubmitVAA(cachedCtx, &wormholetypes.MsgSubmitVAA{
				Vaa: vaa,
			})

			if err == nil {
				writeCache()
			} else {
				ctx.Logger().Info("failed to submit VAA", "err", err)
			}
		}

		return res, nil
	}
}
