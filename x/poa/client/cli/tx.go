package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

// NewTxCmd returns a root CLI command handler for all x/staking transaction commands.
func NewTxCmd() *cobra.Command {
	poaTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "POA transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	poaTxCmd.AddCommand(
		NewCreateValidatorCmd(),
		NewVoteValidatorCmd(),
	)

	return poaTxCmd
}

func NewCreateValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).
				WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			msg, err := newBuildCreateValidatorMsg(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetPublicKey())
	cmd.Flags().AddFlagSet(flagSetDescriptionCreate())

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	_ = cmd.MarkFlagRequired(FlagPubKey)
	_ = cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

func newBuildCreateValidatorMsg(clientCtx client.Context, fs *flag.FlagSet) (*types.MsgCreateValidator, error) {
	valAddr := clientCtx.GetFromAddress()

	pkStr, err := fs.GetString(FlagPubKey)
	if err != nil {
		return nil, err
	}

	var valPubKey []byte

	if strings.HasPrefix(pkStr, sdk.Bech32PrefixConsPub) {
		valPubKey, err = sdk.GetFromBech32(sdk.Bech32PrefixConsPub, pkStr)
		if err != nil {
			return nil, err
		}
	} else {
		var pk cryptotypes.PubKey
		if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &pk); err != nil {
			return nil, err
		}
		valPubKey = pk.Bytes()
	}

	pubKeyAny, err := cdctypes.NewAnyWithValue(&crypto.PublicKey{Sum: &crypto.PublicKey_Ed25519{Ed25519: valPubKey}})
	if err != nil {
		return nil, err

	}

	moniker, _ := fs.GetString(FlagMoniker)
	identity, _ := fs.GetString(FlagIdentity)
	website, _ := fs.GetString(FlagWebsite)
	security, _ := fs.GetString(FlagSecurityContact)
	details, _ := fs.GetString(FlagDetails)
	description := stakingtypes.NewDescription(
		moniker,
		identity,
		website,
		security,
		details,
	)

	msg := &types.MsgCreateValidator{
		Description: description,
		Address:     valAddr.String(),
		Pubkey:      pubKeyAny,
	}

	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Return the flagset, particular flags, and a description of defaults
// this is anticipated to be used with the gen-tx
func CreateValidatorMsgFlagSet(ipDefault string) *flag.FlagSet {
	fsCreateValidator := flag.NewFlagSet("", flag.ContinueOnError)
	fsCreateValidator.String(FlagIP, ipDefault, "The node's public IP")
	fsCreateValidator.String(FlagNodeID, "", "The node's NodeID")
	fsCreateValidator.String(FlagMoniker, "", "The validator's (optional) moniker")
	fsCreateValidator.String(FlagWebsite, "", "The validator's (optional) website")
	fsCreateValidator.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fsCreateValidator.String(FlagDetails, "", "The validator's (optional) details")
	fsCreateValidator.String(FlagIdentity, "", "The (optional) identity signature (ex. UPort or Keybase)")
	fsCreateValidator.AddFlagSet(FlagSetPublicKey())

	return fsCreateValidator
}

func newBuildVoteValidatorMsg(clientCtx client.Context, candidate string, inFavor bool) (*types.MsgVoteValidator, error) {
	valAddr := clientCtx.GetFromAddress()

	msg := &types.MsgVoteValidator{
		CandidateAddress: candidate,
		VoterAddress:     valAddr.String(),
		InFavor:          inFavor,
	}

	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func NewVoteValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-validator [candidate-address] [in-favor]",
		Args:  cobra.ExactArgs(2),
		Short: "vote for a candidate validator (y/n)",
		Example: `tx poa vote-validator <validator-bech32> yes
		tx poa vote-validator <validator-bech32> no		
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).
				WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			inf := strings.ToLower(args[1])

			inFavor := inf == "yes" || inf == "true" || inf == "y" || inf == "1"
			notInFavor := inf == "no" || inf == "false" || inf == "n" || inf == "0"

			if !inFavor && !notInFavor {
				return fmt.Errorf("no valid option provided for [in-favor]. valid: (y,n,yes,no,true,false,1,0)")
			}

			msg, err := newBuildVoteValidatorMsg(clientCtx, args[0], inFavor)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}
