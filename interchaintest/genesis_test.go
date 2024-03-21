package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"

	simappparams "cosmossdk.io/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/icza/dyno"
	tokenfactorytypes "github.com/noble-assets/noble/v5/x/tokenfactory/types"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/relayer"
	"github.com/strangelove-ventures/interchaintest/v4/relayer/rly"
	proposaltypes "github.com/strangelove-ventures/paramauthority/x/params/types/proposal"
	upgradetypes "github.com/strangelove-ventures/paramauthority/x/upgrade/types"
)

var nobleImageInfo = []ibc.DockerImage{
	{
		Repository: "noble",
		Version:    "local",
		UidGid:     "1025:1025",
	},
}

var (
	denomMetadataFrienzies = DenomMetadata{
		Display: "ufrienzies",
		Base:    "ufrienzies",
		Name:    "frienzies",
		Symbol:  "FRNZ",
		DenomUnits: []DenomUnit{
			{
				Denom: "ufrienzies",
				Aliases: []string{
					"microfrienzies",
				},
				Exponent: "0",
			},
			{
				Denom: "mfrienzies",
				Aliases: []string{
					"millifrienzies",
				},
				Exponent: "3",
			},
			{
				Denom:    "frienzies",
				Exponent: "6",
			},
		},
	}

	denomMetadataUsdc = DenomMetadata{
		Display: "usdc",
		Name:    "usdc",
		Base:    "uusdc",
		DenomUnits: []DenomUnit{
			{
				Denom: "uusdc",
				Aliases: []string{
					"microusdc",
				},
				Exponent: "0",
			},
			{
				Denom:    "usdc",
				Exponent: "6",
			},
		},
	}

	defaultShare                   = "0.8"
	defaultDistributionEntityShare = "1.0"
	defaultTransferBPSFee          = "1"
	defaultTransferMaxFee          = "5000000"
	defaultTransferFeeDenom        = denomMetadataUsdc.Base

	relayerImage = relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "v2.4.2", rly.RlyDefaultUidGid)
)

type DenomMetadata struct {
	Display    string      `json:"display"`
	Base       string      `json:"base"`
	Name       string      `json:"name"`
	Symbol     string      `json:"symbol"`
	DenomUnits []DenomUnit `json:"denom_units"`
}

type DenomUnit struct {
	Denom    string   `json:"denom"`
	Aliases  []string `json:"aliases"`
	Exponent string   `json:"exponent"`
}

type TokenFactoryAddress struct {
	Address string `json:"address"`
}

type ParamAuthAddress struct {
	Address string `json:"address"`
}

type TokenFactoryPaused struct {
	Paused bool `json:"paused"`
}

type TokenFactoryDenom struct {
	Denom string `json:"denom"`
}

type DistributionEntity struct {
	Address string `json:"address"`
	Share   string `json:"share"`
}

type CCTPAmount struct {
	Amount string `json:"amount"`
}

type CCTPPerMessageBurnLimit struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

type CCTPNumber struct {
	Amount string `json:"amount"`
}

type CCTPNonce struct {
	Nonce string `json:"nonce"`
}

type Attester struct {
	Attester string `json:"attester"`
}

func NobleEncoding() *simappparams.EncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	proposaltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	upgradetypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

type ExtraWallets struct {
	User  ibc.Wallet
	User2 ibc.Wallet
	Alice ibc.Wallet
}

type NobleRoles struct {
	Owner             ibc.Wallet
	Owner2            ibc.Wallet
	MasterMinter      ibc.Wallet
	MinterController  ibc.Wallet
	MinterController2 ibc.Wallet
	Minter            ibc.Wallet
	Blacklister       ibc.Wallet
	Pauser            ibc.Wallet
}

// Creates tokenfactory wallets. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createTokenfactoryRoles(ctx context.Context, denomMetadata DenomMetadata, val *cosmos.ChainNode, minSetup bool) (NobleRoles, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error

	nobleRoles := NobleRoles{}

	nobleRoles.Owner, err = nobleVal.BuildRelayerWallet(ctx, "owner-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	if err := val.RecoverKey(ctx, nobleRoles.Owner.KeyName(), nobleRoles.Owner.Mnemonic()); err != nil {
		return NobleRoles{}, fmt.Errorf("failed to restore %s wallet: %w", nobleRoles.Owner.KeyName(), err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: nobleRoles.Owner.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  0,
	}
	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []types.Coin{types.NewCoin(genesisWallet.Denom, types.NewIntFromUint64(uint64(genesisWallet.Amount)))})
	if err != nil {
		return NobleRoles{}, err
	}
	if minSetup {
		return nobleRoles, nil
	}

	nobleRoles.Owner2, err = nobleVal.BuildRelayerWallet(ctx, "owner2-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "owner2", err)
	}
	nobleRoles.MasterMinter, err = nobleVal.BuildRelayerWallet(ctx, "masterminter-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "masterminter", err)
	}
	nobleRoles.MinterController, err = nobleVal.BuildRelayerWallet(ctx, "mintercontroller-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "mintercontroller", err)
	}
	nobleRoles.MinterController2, err = nobleVal.BuildRelayerWallet(ctx, "mintercontroller2-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "mintercontroller2", err)
	}
	nobleRoles.Minter, err = nobleVal.BuildRelayerWallet(ctx, "minter-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "minter", err)
	}
	nobleRoles.Blacklister, err = nobleVal.BuildRelayerWallet(ctx, "blacklister-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "blacklister", err)
	}
	nobleRoles.Pauser, err = nobleVal.BuildRelayerWallet(ctx, "pauser-"+denomMetadata.Base)
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "pauser", err)
	}

	walletsToRestore := []ibc.Wallet{nobleRoles.Owner2, nobleRoles.MasterMinter, nobleRoles.MinterController, nobleRoles.MinterController2, nobleRoles.Minter, nobleRoles.Blacklister, nobleRoles.Pauser}
	for _, wallet := range walletsToRestore {
		if err = val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return NobleRoles{}, fmt.Errorf("failed to restore %s wallet: %w", wallet.KeyName(), err)
		}
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: nobleRoles.Owner2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MasterMinter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MinterController.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MinterController2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Minter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Blacklister.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Pauser.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, types.NewIntFromUint64(uint64(wallet.Amount)))})
		if err != nil {
			return NobleRoles{}, err
		}
	}

	return nobleRoles, nil
}

// Creates extra wallets used for testing. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createParamAuthAtGenesis(ctx context.Context, val *cosmos.ChainNode) (ibc.Wallet, error) {
	chainCfg := val.Chain.Config()

	// Test address: noble127de05h6z3a3rh5jf0rjepa48zpgxtesfywgtf
	wallet, err := val.Chain.BuildWallet(ctx, "authority", "index grain inform faith cave know pluck avoid supply zoo retreat system perfect aware shuffle abuse fat security cash amount night return grape candy")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: wallet.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  0,
	}

	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []types.Coin{types.NewCoin(genesisWallet.Denom, types.NewIntFromUint64(uint64(genesisWallet.Amount)))})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

// Creates extra wallets used for testing. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createExtraWalletsAtGenesis(ctx context.Context, val *cosmos.ChainNode) (ExtraWallets, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error

	extraWallets := &ExtraWallets{}

	extraWallets.User, err = nobleVal.BuildRelayerWallet(ctx, "user")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	extraWallets.User2, err = nobleVal.BuildRelayerWallet(ctx, "user2")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	extraWallets.Alice, err = nobleVal.BuildRelayerWallet(ctx, "alice")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	walletsToRestore := []ibc.Wallet{extraWallets.User, extraWallets.User2, extraWallets.Alice}
	for _, wallet := range walletsToRestore {
		if err = val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return ExtraWallets{}, fmt.Errorf("failed to restore %s wallet: %w", wallet.KeyName(), err)
		}
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: extraWallets.User.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: extraWallets.User2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: extraWallets.Alice.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, types.NewIntFromUint64(uint64(wallet.Amount)))})
		if err != nil {
			return ExtraWallets{}, err
		}
	}
	return *extraWallets, nil
}

type genesisWrapper struct {
	chain          *cosmos.CosmosChain
	tfRoles        NobleRoles
	fiatTfRoles    NobleRoles
	paramAuthority ibc.Wallet
	extraWallets   ExtraWallets
}

func nobleChainSpec(
	ctx context.Context,
	gw *genesisWrapper,
	chainID string,
	nv, nf int,
	minSetupTf, minSetupFiatTf bool,
	minModifyTf, minModifyFiatTf bool,
) *interchaintest.ChainSpec {
	return &interchaintest.ChainSpec{
		NumValidators: &nv,
		NumFullNodes:  &nf,
		ChainConfig: ibc.ChainConfig{
			Type:           "cosmos",
			Name:           "noble",
			ChainID:        chainID,
			Bin:            "nobled",
			Denom:          "token",
			Bech32Prefix:   "noble",
			CoinType:       "118",
			GasPrices:      "0.0token",
			GasAdjustment:  1.1,
			TrustingPeriod: "504h",
			NoHostMount:    false,
			Images:         nobleImageInfo,
			EncodingConfig: NobleEncoding(),
			PreGenesis:     preGenesisAll(ctx, gw, minSetupTf, minSetupFiatTf),
			ModifyGenesis:  modifyGenesisAll(gw, minModifyTf, minModifyFiatTf),
		},
	}
}

func preGenesisAll(ctx context.Context, gw *genesisWrapper, minSetupTf, minSetupFiatTf bool) func(ibc.ChainConfig) error {
	return func(cc ibc.ChainConfig) (err error) {
		val := gw.chain.Validators[0]

		gw.tfRoles, err = createTokenfactoryRoles(ctx, denomMetadataFrienzies, val, minSetupTf)
		if err != nil {
			return err
		}

		gw.fiatTfRoles, err = createTokenfactoryRoles(ctx, denomMetadataUsdc, val, minSetupFiatTf)
		if err != nil {
			return err
		}

		gw.extraWallets, err = createExtraWalletsAtGenesis(ctx, val)
		if err != nil {
			return err
		}

		gw.paramAuthority, err = createParamAuthAtGenesis(ctx, val)
		return err
	}
}

func modifyGenesisAll(gw *genesisWrapper, minSetupTf, minSetupFiatTf bool) func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
	return func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		g := make(map[string]interface{})

		if err := json.Unmarshal(b, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}

		if err := modifyGenesisTokenfactory(g, "tokenfactory", denomMetadataFrienzies, gw.tfRoles, minSetupTf); err != nil {
			return nil, err
		}

		if err := modifyGenesisTokenfactory(g, "fiat-tokenfactory", denomMetadataUsdc, gw.fiatTfRoles, minSetupFiatTf); err != nil {
			return nil, err
		}

		authority := gw.paramAuthority.FormattedAddress()

		if err := modifyGenesisParamAuthority(g, authority); err != nil {
			return nil, err
		}

		if err := modifyGenesisTariffDefaults(g, authority); err != nil {
			return nil, err
		}

		if err := modifyGenesisCCTP(g, gw.fiatTfRoles.Owner.FormattedAddress()); err != nil {
			return nil, err
		}

		out, err := json.Marshal(&g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}

		return out, nil
	}
}

// Modifies tokenfactory genesis accounts.
// If minSetup = true, only the owner address, paused state, and denom is setup in genesis.
// These are minimum requirements to start the chain. Otherwise all tokenfactory accounts are created.
func modifyGenesisTokenfactory(g map[string]interface{}, tokenfactoryModName string, denomMetadata DenomMetadata, roles NobleRoles, minSetup bool) error {
	if err := dyno.Set(g, TokenFactoryAddress{roles.Owner.FormattedAddress()}, "app_state", tokenfactoryModName, "owner"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryPaused{false}, "app_state", tokenfactoryModName, "paused"); err != nil {
		return fmt.Errorf("failed to set paused in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryDenom{denomMetadata.Base}, "app_state", tokenfactoryModName, "mintingDenom"); err != nil {
		return fmt.Errorf("failed to set minting denom in genesis json: %w", err)
	}
	if err := dyno.Append(g, denomMetadata, "app_state", "bank", "denom_metadata"); err != nil {
		return fmt.Errorf("failed to set denom metadata in genesis json: %w", err)
	}
	if minSetup {
		return nil
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.MasterMinter.FormattedAddress()}, "app_state", tokenfactoryModName, "masterMinter"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Blacklister.FormattedAddress()}, "app_state", tokenfactoryModName, "blacklister"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Pauser.FormattedAddress()}, "app_state", tokenfactoryModName, "pauser"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	return nil
}

func modifyGenesisParamAuthority(genbz map[string]interface{}, authorityAddress string) error {
	if err := dyno.Set(genbz, authorityAddress, "app_state", "params", "params", "authority"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, authorityAddress, "app_state", "upgrade", "params", "authority"); err != nil {
		return fmt.Errorf("failed to set upgrade authority address in genesis json: %w", err)
	}

	return nil
}

// "params": {},
// "authority": "our address",
// "attester_list": [],
// "per_message_burn_limit_list": [],
// "burning_and_minting_paused": false,
// "sending_and_receiving_messages_paused": false,
// "max_message_body_size": 8000,
// "next_available_nonce": 0,
// "signature_threshold": 2,
// "token_pair_list": [],
// "used_nonces_list": []
// "token_messenger_list": []
func modifyGenesisCCTP(genbz map[string]interface{}, authority string) error {
	if err := dyno.Set(genbz, authority, "app_state", "cctp", "owner"); err != nil {
		return fmt.Errorf("failed to set cctp authority address in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, authority, "app_state", "cctp", "attester_manager"); err != nil {
		return fmt.Errorf("failed to set cctp authority address in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, authority, "app_state", "cctp", "pauser"); err != nil {
		return fmt.Errorf("failed to set cctp authority address in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, authority, "app_state", "cctp", "token_controller"); err != nil {
		return fmt.Errorf("failed to set cctp authority address in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, []CCTPPerMessageBurnLimit{{Amount: "99999999", Denom: denomMetadataUsdc.Base}}, "app_state", "cctp", "per_message_burn_limit_list"); err != nil {
		return fmt.Errorf("failed to set cctp perMessageBurnLimit in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, CCTPAmount{Amount: "8000"}, "app_state", "cctp", "max_message_body_size"); err != nil {
		return fmt.Errorf("failed to set cctp maxMessageBodySize in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, CCTPNonce{Nonce: "0"}, "app_state", "cctp", "next_available_nonce"); err != nil {
		return fmt.Errorf("failed to set cctp nonce in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, CCTPAmount{Amount: "2"}, "app_state", "cctp", "signature_threshold"); err != nil {
		return fmt.Errorf("failed to set cctp signatureThreshold in genesis json: %w", err)
	}
	//if err := dyno.Set(genbz, Attester{Attester: "0xE2fEfe09E74b921CbbFF229E7cD40009231501CA"}, "app_state", "cctp", "attes"); err != nil {
	//	return fmt.Errorf("failed to set cctp signatureThreshold in genesis json: %w", err)
	//}
	//if err := dyno.Set(genbz, Attester{Attester: "0xb0Ea8E1bE37F346C7EA7ec708834D0db18A17361"}, "app_state", "cctp", "attes"); err != nil {
	//	return fmt.Errorf("failed to set cctp signatureThreshold in genesis json: %w", err)
	//}

	return nil
}

func modifyGenesisTariffDefaults(
	genbz map[string]interface{},
	distributionEntity string,
) error {
	return modifyGenesisTariff(genbz, defaultShare, distributionEntity,
		defaultDistributionEntityShare, defaultTransferBPSFee, defaultTransferMaxFee, defaultTransferFeeDenom)
}

func modifyGenesisTariff(
	genbz map[string]interface{},
	share string,
	distributionEntity string,
	distributionEntityShare string,
	transferBPSFee string,
	transferMaxFee string,
	transferDenom string,
) error {
	if err := dyno.Set(genbz, share, "app_state", "tariff", "params", "share"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	distributionEntities := []DistributionEntity{
		{
			Address: distributionEntity,
			Share:   distributionEntityShare,
		},
	}
	if err := dyno.Set(genbz, distributionEntities, "app_state", "tariff", "params", "distribution_entities"); err != nil {
		return fmt.Errorf("failed to set upgrade authority address in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, transferBPSFee, "app_state", "tariff", "params", "transfer_fee_bps"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, transferMaxFee, "app_state", "tariff", "params", "transfer_fee_max"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, transferDenom, "app_state", "tariff", "params", "transfer_fee_denom"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	return nil
}

func modifyGenesisDowntimeWindow(bz map[string]interface{}) error {
	return dyno.Set(bz, "5", "app_state", "slashing", "params", "signed_blocks_window")
}
