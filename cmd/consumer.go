package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/x/ccv/consumer/types"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/testutil"
	types1 "github.com/tendermint/tendermint/abci/types"
	pvm "github.com/tendermint/tendermint/privval"
	tmtypes "github.com/tendermint/tendermint/types"
)

func AddConsumerSectionCmd(defaultNodeHome string) *cobra.Command {
	genesisMutator := NewDefaultGenesisIO()

	txCmd := &cobra.Command{
		Use:                        "add-consumer-section",
		Short:                      "ONLY FOR TESTING PURPOSES! Modifies genesis so that chain can be started locally with one node.",
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			return genesisMutator.AlterConsumerModuleState(cmd, func(state *GenesisData, _ map[string]json.RawMessage) error {
				genesisState := testutil.CreateMinimalConsumerTestGenesis()
				clientCtx := client.GetClientContextFromCmd(cmd)
				serverCtx := server.GetServerContextFromCmd(cmd)
				config := serverCtx.Config
				config.SetRoot(clientCtx.HomeDir)
				privValidator := pvm.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
				pk, err := privValidator.GetPubKey()
				if err != nil {
					return err
				}
				sdkPublicKey, err := cryptocodec.FromTmPubKeyInterface(pk)
				if err != nil {
					return err
				}
				tmProtoPublicKey, err := cryptocodec.ToTmProtoPublicKey(sdkPublicKey)
				if err != nil {
					return err
				}

				initialValset := []types1.ValidatorUpdate{{PubKey: tmProtoPublicKey, Power: 100}}
				vals, err := tmtypes.PB2TM.ValidatorUpdates(initialValset)
				if err != nil {
					return sdkerrors.Wrap(err, "could not convert val updates to validator set")
				}

				genesisState.InitialValSet = initialValset
				genesisState.ProviderConsensusState.NextValidatorsHash = tmtypes.NewValidatorSet(vals).Hash()

				state.ConsumerModuleState = genesisState

				return nil
			})
		},
	}

	txCmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(txCmd)

	return txCmd
}

type GenesisMutator interface {
	AlterConsumerModuleState(cmd *cobra.Command, callback func(state *GenesisData, appState map[string]json.RawMessage) error) error
}

type DefaultGenesisIO struct {
	DefaultGenesisReader
}

func NewDefaultGenesisIO() *DefaultGenesisIO {
	return &DefaultGenesisIO{DefaultGenesisReader: DefaultGenesisReader{}}
}

func (x DefaultGenesisIO) AlterConsumerModuleState(cmd *cobra.Command, callback func(state *GenesisData, appState map[string]json.RawMessage) error) error {
	g, err := x.ReadGenesis(cmd)
	if err != nil {
		return err
	}
	if err := callback(g, g.AppState); err != nil {
		return err
	}
	if err := g.ConsumerModuleState.Validate(); err != nil {
		return err
	}
	clientCtx := client.GetClientContextFromCmd(cmd)
	consumerGenStateBz, err := clientCtx.Codec.MarshalJSON(g.ConsumerModuleState)
	if err != nil {
		return sdkerrors.Wrap(err, "marshal consumer genesis state")
	}

	g.AppState[ccvconsumertypes.ModuleName] = consumerGenStateBz
	appStateJSON, err := json.Marshal(g.AppState)
	if err != nil {
		return sdkerrors.Wrap(err, "marshal application genesis state")
	}

	g.GenDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(g.GenDoc, g.GenesisFile)
}

type DefaultGenesisReader struct{}

func (d DefaultGenesisReader) ReadGenesis(cmd *cobra.Command) (*GenesisData, error) {
	clientCtx := client.GetClientContextFromCmd(cmd)
	serverCtx := server.GetServerContextFromCmd(cmd)
	config := serverCtx.Config
	config.SetRoot(clientCtx.HomeDir)

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	return NewGenesisData(
		genFile,
		genDoc,
		appState,
		nil,
	), nil
}

type GenesisData struct {
	GenesisFile         string
	GenDoc              *tmtypes.GenesisDoc
	AppState            map[string]json.RawMessage
	ConsumerModuleState *ccvconsumertypes.GenesisState
}

func NewGenesisData(genesisFile string, genDoc *tmtypes.GenesisDoc, appState map[string]json.RawMessage, consumerModuleState *ccvconsumertypes.GenesisState) *GenesisData {
	return &GenesisData{GenesisFile: genesisFile, GenDoc: genDoc, AppState: appState, ConsumerModuleState: consumerModuleState}
}
