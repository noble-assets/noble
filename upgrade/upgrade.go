package upgrade

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	v6 "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/migrations/v6"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"
	authoritykeeper "github.com/noble-assets/authority/keeper"
	authoritytypes "github.com/noble-assets/authority/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	cdc codec.Codec,
	logger log.Logger,
	capabilityStoreKey *storetypes.KVStoreKey,
	accountKeeper authkeeper.AccountKeeper,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	capabilityKeeper *capabilitykeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	consensusKeeper consensuskeeper.Keeper,
	paramsKeeper paramskeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// Initialize legacy param subspaces with key tables for migration.
		// NOTE: This must be done before RunMigrations is executed.
		for _, subspace := range paramsKeeper.GetSubspaces() {
			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
			case distributiontypes.ModuleName:
				keyTable = distributiontypes.ParamKeyTable() //nolint:staticcheck
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck

			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Don't run InitGenesis on x/authority module, so we can migrate the
		// legacy ParamAuthority address later.
		vm[authoritytypes.ModuleName] = 1

		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		// ----- IBC Specific Logic -----
		// https://github.com/cosmos/ibc-go/blob/v8.2.1/testing/simapp/upgrades/upgrades.go

		// IBC v5 -> v6: Migrate ICS-27 channel capabilities.
		// https://ibc.cosmos.network/main/migrations/v5-to-v6
		err = v6.MigrateICS27ChannelCapability(sdkCtx, cdc, capabilityStoreKey, capabilityKeeper, icahosttypes.SubModuleName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to migrate ICS-27 channel capabilities")
		}
		// IBC v6 -> v7: Prune the consensus states of expired Tendermint light clients.
		// https://ibc.cosmos.network/main/migrations/v6-to-v7
		_, err = ibctmmigrations.PruneExpiredConsensusStates(sdkCtx, cdc, clientKeeper)
		if err != nil {
			return nil, errors.Wrap(err, "failed to prune expired consensus states")
		}
		// IBC v7 -> v7.1: Register 09-localhost as an allowed light client.
		// https://ibc.cosmos.network/main/migrations/v7-to-v7_1
		params := clientKeeper.GetParams(sdkCtx)
		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		clientKeeper.SetParams(sdkCtx, params)

		// ----- SDK Specific Logic -----
		// https://docs.cosmos.network/main/build/migrations/upgrading

		// SDK v0.46 -> v0.47: Migrate CometBFT params to x/consensus module.
		// https://docs.cosmos.network/main/build/migrations/upgrading#xconsensus
		subspace := paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()) //nolint:staticcheck
		err = baseapp.MigrateParams(sdkCtx, subspace, consensusKeeper.ParamsStore)
		if err != nil {
			return nil, errors.Wrap(err, "failed to migrate consensus params")
		}

		// ----- Noble Specific Logic -----

		// Migrate ParamAuthority address to x/authority module.
		var authority string
		subspace = paramsKeeper.Subspace(paramstypes.ModuleName).WithKeyTable(authoritytypes.ParamKeyTable()) //nolint:staticcheck
		subspace.Get(sdkCtx, authoritytypes.AuthorityKey, &authority)
		err = authorityKeeper.Owner.Set(ctx, authority)
		if err != nil {
			return vm, errors.Wrap(err, "failed to migrate authority address")
		}

		MigrateValidatorAccounts(ctx, accountKeeper, stakingKeeper)
		err = BurnSurplusSupply(ctx, authority, accountKeeper, bankKeeper)
		if err != nil {
			return vm, err
		}

		logger.Info(UpgradeASCII)
		return vm, nil
	}
}

// MigrateValidatorAccounts performs a migration of all validator operators to
// permanently locked vesting accounts. NOTE: In a future upgrade, think about
// clawing back inactive validator staking tokens.
func MigrateValidatorAccounts(ctx context.Context, accountKeeper authkeeper.AccountKeeper, stakingKeeper *stakingkeeper.Keeper) {
	validators, _ := stakingKeeper.GetAllValidators(ctx)
	for _, validator := range validators {
		operator, _ := stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.OperatorAddress)
		rawAccount := accountKeeper.GetAccount(ctx, operator)

		switch account := rawAccount.(type) {
		case *vestingtypes.ContinuousVestingAccount:
			rawAccount = &vestingtypes.PermanentLockedAccount{
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:      account.BaseAccount,
					OriginalVesting:  account.OriginalVesting,
					DelegatedFree:    sdk.NewCoins(),
					DelegatedVesting: account.OriginalVesting,
					EndTime:          0,
				},
			}
		case *vestingtypes.DelayedVestingAccount:
			rawAccount = &vestingtypes.PermanentLockedAccount{
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:      account.BaseAccount,
					OriginalVesting:  account.OriginalVesting,
					DelegatedFree:    sdk.NewCoins(),
					DelegatedVesting: account.OriginalVesting,
					EndTime:          0,
				},
			}
		}

		accountKeeper.SetAccount(ctx, rawAccount)
	}
}

// BurnSurplusSupply performs a burn of the surplus $STAKE supply.
func BurnSurplusSupply(ctx context.Context, authority string, accountKeeper authkeeper.AccountKeeper, bankKeeper bankkeeper.Keeper) error {
	supply := bankKeeper.GetSupply(ctx, "ustake")
	surplus := supply.Sub(sdk.NewCoin(
		"ustake", math.NewInt(1_000_000_000_000_000),
	))

	if !surplus.IsPositive() {
		return nil
	}

	address, err := accountKeeper.AddressCodec().StringToBytes(authority)
	if err != nil {
		return err
	}
	err = bankKeeper.SendCoinsFromAccountToModule(ctx, address, upgradetypes.ModuleName, sdk.NewCoins(surplus))
	if err != nil {
		return err
	}

	err = bankKeeper.BurnCoins(ctx, upgradetypes.ModuleName, sdk.NewCoins(surplus))
	return err
}
