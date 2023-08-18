package app

import (
	"time"

	runtime "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	app "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/appconfig"
	"google.golang.org/protobuf/types/known/durationpb"

	// Auth
	auth "cosmossdk.io/api/cosmos/auth/module/v1"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// Authority
	authority "github.com/noble-assets/paramauthority/pulsar/noble/authority/module/v1"
	authorityTypes "github.com/noble-assets/paramauthority/x/authority/types"
	// Authz
	authz "cosmossdk.io/api/cosmos/authz/module/v1"
	authzTypes "github.com/cosmos/cosmos-sdk/x/authz"
	// Bank
	bank "cosmossdk.io/api/cosmos/bank/module/v1"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	// Capability
	capability "cosmossdk.io/api/cosmos/capability/module/v1"
	capabilityTypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	// Consensus
	consensus "cosmossdk.io/api/cosmos/consensus/module/v1"
	consensusTypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	// Consumer
	consumerTypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	// Crisis
	crisis "cosmossdk.io/api/cosmos/crisis/module/v1"
	crisisTypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	// Distribution
	distribution "cosmossdk.io/api/cosmos/distribution/module/v1"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	// Evidence
	evidence "cosmossdk.io/api/cosmos/evidence/module/v1"
	evidenceTypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	// FeeGrant
	feeGrant "cosmossdk.io/api/cosmos/feegrant/module/v1"
	feeGrantTypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	// FiatTokenFactory
	fiatTokenFactory "github.com/circlefin/noble-fiattokenfactory/pulsar/circle/fiattokenfactory/module/v1"
	fiatTokenFactoryTypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	// GenUtil
	genUtil "cosmossdk.io/api/cosmos/genutil/module/v1"
	genUtilTypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	// GlobalFee
	globalFee "github.com/strangelove-ventures/noble/pulsar/noble/globalfee/module/v1"
	globalFeeTypes "github.com/strangelove-ventures/noble/x/globalfee/types"
	// Group
	group "cosmossdk.io/api/cosmos/group/module/v1"
	groupTypes "github.com/cosmos/cosmos-sdk/x/group"
	// IBC Core
	ibcTypes "github.com/cosmos/ibc-go/v7/modules/core/exported"
	// IBC Fee
	ibcFeeTypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	// IBC Transfer
	ibcTransferTypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	// ICA
	icaTypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	// Params
	params "cosmossdk.io/api/cosmos/params/module/v1"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	// PFM
	pfmTypes "github.com/strangelove-ventures/packet-forward-middleware/v7/router/types"
	// Slashing
	slashing "cosmossdk.io/api/cosmos/slashing/module/v1"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	// Staking
	staking "cosmossdk.io/api/cosmos/staking/module/v1"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	// Tariff
	tariff "github.com/strangelove-ventures/noble/pulsar/noble/tariff/module/v1"
	tariffTypes "github.com/strangelove-ventures/noble/x/tariff/types"
	// TokenFactory
	tokenFactory "github.com/strangelove-ventures/noble/pulsar/noble/tokenfactory/module/v1"
	tokenFactoryTypes "github.com/strangelove-ventures/noble/x/tokenfactory/types"
	// Tx
	tx "cosmossdk.io/api/cosmos/tx/config/v1"
	// Upgrade
	upgrade "cosmossdk.io/api/cosmos/upgrade/module/v1"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	// Vesting
	vesting "cosmossdk.io/api/cosmos/vesting/module/v1"
	vestingTypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var AppConfig = appconfig.Compose(&app.Config{
	Modules: []*app.ModuleConfig{
		{
			Name: "runtime",
			Config: appconfig.WrapAny(&runtime.Module{
				AppName: "Noble",
				BeginBlockers: []string{
					upgradeTypes.ModuleName,
					capabilityTypes.ModuleName,
					tariffTypes.ModuleName, // Noble
					distributionTypes.ModuleName,
					slashingTypes.ModuleName,
					evidenceTypes.ModuleName,
					stakingTypes.ModuleName,
					authTypes.ModuleName,
					bankTypes.ModuleName,
					crisisTypes.ModuleName,
					consumerTypes.ModuleName,    // TODO -- ICS
					ibcTransferTypes.ModuleName, // IBC
					ibcTypes.ModuleName,         // IBC
					icaTypes.ModuleName,         // IBC
					ibcFeeTypes.ModuleName,      // IBC
					pfmTypes.ModuleName,         // IBC
					genUtilTypes.ModuleName,
					authzTypes.ModuleName,
					feeGrantTypes.ModuleName,
					groupTypes.ModuleName,
					paramsTypes.ModuleName,
					vestingTypes.ModuleName,
					consensusTypes.ModuleName,
					// Noble
					authorityTypes.ModuleName,
					tokenFactoryTypes.ModuleName,
					fiatTokenFactoryTypes.ModuleName,
					globalFeeTypes.ModuleName,
				},
				EndBlockers: []string{
					crisisTypes.ModuleName,
					stakingTypes.ModuleName,
					consumerTypes.ModuleName,    // TODO -- ICS
					ibcTransferTypes.ModuleName, // IBC
					ibcTypes.ModuleName,         // IBC
					icaTypes.ModuleName,         // IBC
					ibcFeeTypes.ModuleName,      // IBC
					pfmTypes.ModuleName,         // IBC
					capabilityTypes.ModuleName,
					authTypes.ModuleName,
					bankTypes.ModuleName,
					distributionTypes.ModuleName,
					slashingTypes.ModuleName,
					genUtilTypes.ModuleName,
					evidenceTypes.ModuleName,
					authzTypes.ModuleName,
					feeGrantTypes.ModuleName,
					groupTypes.ModuleName,
					paramsTypes.ModuleName,
					consensusTypes.ModuleName,
					upgradeTypes.ModuleName,
					vestingTypes.ModuleName,
					// Noble
					authorityTypes.ModuleName,
					tokenFactoryTypes.ModuleName,
					fiatTokenFactoryTypes.ModuleName,
					globalFeeTypes.ModuleName,
					tariffTypes.ModuleName,
				},
				OverrideStoreKeys: []*runtime.StoreKeyConfig{
					{
						ModuleName: authTypes.ModuleName,
						KvStoreKey: "acc",
					},
				},
				InitGenesis: []string{
					capabilityTypes.ModuleName,
					authTypes.ModuleName,
					bankTypes.ModuleName,
					tariffTypes.ModuleName, // Noble
					distributionTypes.ModuleName,
					stakingTypes.ModuleName,
					slashingTypes.ModuleName,
					crisisTypes.ModuleName,
					genUtilTypes.ModuleName,
					evidenceTypes.ModuleName,
					authzTypes.ModuleName,
					feeGrantTypes.ModuleName,
					groupTypes.ModuleName,
					paramsTypes.ModuleName,
					upgradeTypes.ModuleName,
					vestingTypes.ModuleName,
					consensusTypes.ModuleName,

					consumerTypes.ModuleName, // TODO -- ICS
					ibcTransferTypes.ModuleName,
					ibcTypes.ModuleName,
					icaTypes.ModuleName,
					ibcFeeTypes.ModuleName,
					pfmTypes.ModuleName,

					authorityTypes.ModuleName,
					tokenFactoryTypes.ModuleName,
					fiatTokenFactoryTypes.ModuleName,
					globalFeeTypes.ModuleName,
				},
			}),
		},

		// Cosmos SDK Modules

		{
			Name: authTypes.ModuleName,
			Config: appconfig.WrapAny(&auth.Module{
				Authority:    "authority",
				Bech32Prefix: "noble",
				ModuleAccountPermissions: []*auth.ModuleAccountPermission{
					{Account: authTypes.FeeCollectorName},
					{Account: distributionTypes.ModuleName},
					{Account: stakingTypes.BondedPoolName, Permissions: []string{authTypes.Burner, stakingTypes.ModuleName}},
					{Account: stakingTypes.NotBondedPoolName, Permissions: []string{authTypes.Burner, stakingTypes.ModuleName}},

					{Account: ibcFeeTypes.ModuleName},
					{Account: ibcTransferTypes.ModuleName, Permissions: []string{authTypes.Burner, authTypes.Minter}},
					{Account: icaTypes.ModuleName},

					{Account: fiatTokenFactoryTypes.ModuleName, Permissions: []string{authTypes.Burner, authTypes.Minter, stakingTypes.ModuleName}},
					{Account: tokenFactoryTypes.ModuleName, Permissions: []string{authTypes.Burner, authTypes.Minter, stakingTypes.ModuleName}},
				},
			}),
		},
		{
			Name:   authzTypes.ModuleName,
			Config: appconfig.WrapAny(&authz.Module{}),
		},
		{
			Name: bankTypes.ModuleName,
			Config: appconfig.WrapAny(&bank.Module{
				Authority: "authority",
				BlockedModuleAccountsOverride: []string{
					authTypes.FeeCollectorName,
					distributionTypes.ModuleName,
					stakingTypes.BondedPoolName,
					stakingTypes.NotBondedPoolName,

					authorityTypes.ModuleName,
				},
			}),
		},
		{
			Name: capabilityTypes.ModuleName,
			Config: appconfig.WrapAny(&capability.Module{
				SealKeeper: true,
			}),
		},
		{
			Name: consensusTypes.ModuleName,
			Config: appconfig.WrapAny(&consensus.Module{
				Authority: "authority",
			}),
		},
		{
			Name: crisisTypes.ModuleName,
			Config: appconfig.WrapAny(&crisis.Module{
				Authority: "authority",
			}),
		},
		{
			Name: distributionTypes.ModuleName,
			Config: appconfig.WrapAny(&distribution.Module{
				Authority: "authority",
			}),
		},
		{
			Name:   evidenceTypes.ModuleName,
			Config: appconfig.WrapAny(&evidence.Module{}),
		},
		{
			Name:   feeGrantTypes.ModuleName,
			Config: appconfig.WrapAny(&feeGrant.Module{}),
		},
		{
			Name:   genUtilTypes.ModuleName,
			Config: appconfig.WrapAny(&genUtil.Module{}),
		},
		{
			Name: groupTypes.ModuleName,
			Config: appconfig.WrapAny(&group.Module{
				MaxExecutionPeriod: durationpb.New(time.Second * 1209600),
				MaxMetadataLen:     255,
			}),
		},
		{
			Name:   paramsTypes.ModuleName,
			Config: appconfig.WrapAny(&params.Module{}),
		},
		{
			Name: slashingTypes.ModuleName,
			Config: appconfig.WrapAny(&slashing.Module{
				Authority: "authority",
			}),
		},
		{
			Name: stakingTypes.ModuleName,
			Config: appconfig.WrapAny(&staking.Module{
				Authority: "authority",
			}),
		},
		{
			Name:   "tx",
			Config: appconfig.WrapAny(&tx.Config{}),
		},
		{
			Name: upgradeTypes.ModuleName,
			Config: appconfig.WrapAny(&upgrade.Module{
				Authority: "authority",
			}),
		},
		{
			Name:   vestingTypes.ModuleName,
			Config: appconfig.WrapAny(&vesting.Module{}),
		},

		// Custom Modules

		{
			Name:   authorityTypes.ModuleName,
			Config: appconfig.WrapAny(&authority.Module{}),
		},
		{
			Name:   fiatTokenFactoryTypes.ModuleName,
			Config: appconfig.WrapAny(&fiatTokenFactory.Module{}),
		},
		{
			Name: globalFeeTypes.ModuleName,
			Config: appconfig.WrapAny(&globalFee.Module{
				Authority: "authority",
			}),
		},
		{
			Name: tariffTypes.ModuleName,
			Config: appconfig.WrapAny(&tariff.Module{
				Authority: "authority",
			}),
		},
		{
			Name:   tokenFactoryTypes.ModuleName,
			Config: appconfig.WrapAny(&tokenFactory.Module{}),
		},
	},
})
