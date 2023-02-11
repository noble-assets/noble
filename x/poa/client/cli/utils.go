package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
)

// // CreateValidatorMsgHelpers - used for gen-tx
func CreateValidatorMsgHelpers(ipDefault string) (
	fs *flag.FlagSet, nodeIDFlag, pubkeyFlag, amountFlag, defaultsDesc string) {
	viper.Set(FlagIP, ipDefault)

	fs = flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMoniker, "", "The validator's name")
	fs.String(FlagWebsite, "", "The validator's (optional) website")
	fs.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fs.String(FlagDetails, "", "The validator's (optional) details")
	fs.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	return fs, FlagNodeID, FlagPubKey, "nil", "nil"
}

// // PrepareFlagsForTxCreateValidator - used for gen-tx
func PrepareFlagsForTxCreateValidator(config *cfg.Config, nodeID,
	chainID string, valPubKey crypto.PubKey) {
	viper.Set(flags.FlagChainID, chainID)
	viper.Set(flags.FlagFrom, viper.GetString(flags.FlagName))
	viper.Set(FlagPubKey, sdk.MustBech32ifyAddressBytes(sdk.Bech32PrefixConsPub, valPubKey.Bytes()))
	viper.Set(FlagNodeID, nodeID)
}

// // BuildCreateValidatorMsg - used for gen-tx
func BuildCreateValidatorMsg(cliCtx context.CLIContext,
	txBldr authtypes.TxBuilder) (authtypes.TxBuilder, sdk.Msg, error) {
	pkStr := viper.GetString(flagPubKey)

	valAddr := cliCtx.GetFromAddress()
	consAddr := sdk.ValAddress(cliCtx.GetFromAddress())

	pk, _ := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pkStr)

	moniker := viper.GetString(flagMoniker)
	identity := viper.GetString(flagIdentity)
	website := viper.GetString(flagWebsite)
	securityContact := viper.GetString(flagSecurityContact)
	details := viper.GetString(flagDetails)

	msg := types.NewMsgCreateValidatorPOA(
		consAddr.String(),
		consAddr,
		pk,
		stakingtypes.NewDescription(moniker, identity, website, securityContact, details),
		valAddr,
	)
	ip := viper.GetString(flagIP)
	nodeID := viper.GetString(flagNodeId)

	txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))

	return txBldr, msg, nil
}
