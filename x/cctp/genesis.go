package cctp

import (
	_ "github.com/cosmos/cosmos-sdk/types/errors" // sdkerrors
	"github.com/strangelove-ventures/noble/x/cctp/keeper"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	if genState.Authority != nil {
		k.SetAuthority(ctx, *genState.Authority)
	}

	for _, elem := range genState.PublicKeysList {
		k.SetPublicKey(ctx, elem)
	}

	for _, elem := range genState.MinterAllowanceList {
		k.SetMinterAllowance(ctx, elem)
	}

	if genState.PerMessageBurnLimit != nil {
		k.SetPerMessageBurnLimit(ctx, *genState.PerMessageBurnLimit)
	}

	if genState.BurningAndMintingPaused != nil {
		k.SetBurningAndMintingPaused(ctx, *genState.BurningAndMintingPaused)
	}

	if genState.SendingAndReceivingMessagesPaused != nil {
		k.SetSendingAndReceivingMessagesPaused(ctx, *genState.SendingAndReceivingMessagesPaused)
	}

	if genState.MaxMessageBodySize != nil {
		k.SetMaxMessageBodySize(ctx, *genState.MaxMessageBodySize)
	}

	if genState.Nonce != nil {
		k.SetNonce(ctx, *genState.Nonce)
	} else {
		nonce := types.Nonce{Nonce: 0}
		k.SetNonce(ctx, nonce)
	}

	if genState.SignatureThreshold != nil {
		k.SetSignatureThreshold(ctx, *genState.SignatureThreshold)
	}

	for _, elem := range genState.TokenPairList {
		k.SetTokenPair(ctx, elem)
	}

	for _, elem := range genState.UsedNoncesList {
		k.SetUsedNonce(ctx, keeper.UsedNonce{
			// TODO add to proto
			// SourceDomain: elem.SourceDomain,
			Nonce: elem.Nonce,
		})
	}

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported GenesisState
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	authority, found := k.GetAuthority(ctx)
	if found {
		genesis.Authority = &authority
	}

	genesis.PublicKeysList = k.GetAllPublicKeys(ctx)

	genesis.MinterAllowanceList = k.GetAllMinterAllowances(ctx)

	perMessageBurnLimit, found := k.GetPerMessageBurnLimit(ctx)
	if found {
		genesis.PerMessageBurnLimit = &perMessageBurnLimit
	}

	burningAndMintingPaused, found := k.GetBurningAndMintingPaused(ctx)
	if found {
		genesis.BurningAndMintingPaused = &burningAndMintingPaused
	}

	sendingAndReceivingMessagesPaused, found := k.GetSendingAndReceivingMessagesPaused(ctx)
	if found {
		genesis.SendingAndReceivingMessagesPaused = &sendingAndReceivingMessagesPaused
	}

	maxMessageBodySize, found := k.GetMaxMessageBodySize(ctx)
	if found {
		genesis.MaxMessageBodySize = &maxMessageBodySize
	}

	nonce, found := k.GetNonce(ctx)
	if found {
		genesis.Nonce = &nonce
	}

	signatureThreshold, found := k.GetSignatureThreshold(ctx)
	if found {
		genesis.SignatureThreshold = &signatureThreshold
	}

	genesis.TokenPairList = k.GetAllTokenPairs(ctx)

	usedNonces := k.GetAllUsedNonces(ctx)
	for _, n := range usedNonces {
		genesis.UsedNoncesList = append(genesis.UsedNoncesList, types.Nonce{
			// TODO add to proto
			// SourceDomain: n.SourceDomain,
			Nonce: n.Nonce,
		})
	}

	return genesis
}
