package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/strangelove-ventures/noble/x/poa/types"
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

type TxCreateValidatorConfig struct {
	ChainID string
	NodeID  string
	Moniker string

	PubKey cryptotypes.PubKey

	IP              string
	Website         string
	SecurityContact string
	Details         string
	Identity        string
}

func PrepareConfigForTxCreateValidator(flagSet *flag.FlagSet, moniker, nodeID, chainID string, valPubKey cryptotypes.PubKey) (TxCreateValidatorConfig, error) {
	c := TxCreateValidatorConfig{}

	ip, err := flagSet.GetString(FlagIP)
	if err != nil {
		return c, err
	}
	if ip == "" {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't retrieve an external IP; "+
			"the tx's memo field will be unset")
	}
	c.IP = ip

	website, err := flagSet.GetString(FlagWebsite)
	if err != nil {
		return c, err
	}
	c.Website = website

	securityContact, err := flagSet.GetString(FlagSecurityContact)
	if err != nil {
		return c, err
	}
	c.SecurityContact = securityContact

	details, err := flagSet.GetString(FlagDetails)
	if err != nil {
		return c, err
	}
	c.SecurityContact = details

	identity, err := flagSet.GetString(FlagIdentity)
	if err != nil {
		return c, err
	}
	c.Identity = identity

	c.NodeID = nodeID
	c.PubKey = valPubKey
	c.Website = website
	c.SecurityContact = securityContact
	c.Details = details
	c.Identity = identity
	c.ChainID = chainID
	c.Moniker = moniker

	return c, nil
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(clientCtx client.Context, config TxCreateValidatorConfig, txBldr tx.Factory, generateOnly bool) (tx.Factory, sdk.Msg, error) {
	valAddr := clientCtx.GetFromAddress()
	description := stakingtypes.NewDescription(
		config.Moniker,
		config.Identity,
		config.Website,
		config.SecurityContact,
		config.Details,
	)

	pubKeyAny, err := cdctypes.NewAnyWithValue(config.PubKey)
	if err != nil {
		return txBldr, nil, err
	}

	msg := &types.MsgCreateValidator{
		Description: description,
		Address:     valAddr.String(),
		Pubkey:      pubKeyAny,
	}

	if err != nil {
		return txBldr, msg, err
	}
	if generateOnly {
		ip := config.IP
		nodeID := config.NodeID

		if nodeID != "" && ip != "" {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}

	return txBldr, msg, nil
}

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", string(sdk.MustSortJSON(out)))

	return err
}
