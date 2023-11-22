# CHANGELOG

<<<<<<< HEAD
## v4.1.0-rc.0

*Nov 15, 2023*

This is the first release candidate for a minor release to the v4 Argon line.

### DEPENDENCIES

- Bump CCTP from [`dc81b3e`](https://github.com/circlefin/noble-cctp/commit/dc81b3e0d566d195c869a213519fcecd38b020a5) to [`86f425e`](https://github.com/circlefin/noble-cctp/commit/86f425e6fac94ff57865dd11b42c95de303e0d96) ([#259](https://github.com/strangelove-ventures/noble/pull/259))

### FEATURES

- Introduce a new `x/stabletokenfactory` module for issuing [USDLR by Stable](https://withstable.com). ([#269](https://github.com/strangelove-ventures/noble/pull/269))

### IMPROVEMENTS

- Align module path with Go's [naming convention](https://go.dev/doc/modules/version-numbers#major-version). ([#249](https://github.com/strangelove-ventures/noble/pull/249))
- Remove `x/fiattokenfactory` module from codebase and switch to migrated version under [`circlefin/noble-fiattokenfactory`](https://github.com/circlefin/noble-fiattokenfactory) ([#259](https://github.com/strangelove-ventures/noble/pull/259))
- Add multiple fee denom support to the `x/tariff` module. ([#269](https://github.com/strangelove-ventures/noble/pull/269))
=======
## v4.0.2

*Nov 21, 2023*

This is a non-consensus breaking patch release to the v4 Argon line.

### IMPROVEMENTS

- Implement a parameter query for the `x/tariff` module. ([#277](https://github.com/noble-assets/noble/pull/277))

## v4.0.1

*Nov 16, 2023*

This is a consensus breaking patch release to the v4 Argon line.

### BUG FIXES

- Unregister `x/distribution` hooks to address consensus failure. ([#274](https://github.com/noble-assets/noble/pull/274))
>>>>>>> a4ad980 (chore: rename module path (#283))

## v4.0.0

*Nov 6, 2023*

This is the long awaited Argon major release of Noble. It introduces a new [`x/cctp`](https://github.com/circlefin/noble-cctp) module that implements Circle's [Cross Chain Transfer Protocol (CCTP)](https://www.circle.com/en/cross-chain-transfer-protocol), allowing native $USDC transfers between supported EVM networks and Noble (with many more networks to come). 

Along with the integration of the CCTP module, the following changes were made.

### BUG FIXES

- Fix simulation tests. ([#252](https://github.com/noble-assets/noble/pull/252))
- Fix Ledger support for macOS Sonoma. ([#253](https://github.com/noble-assets/noble/pull/253))

### DEPENDENCIES

- Bump IBC to [`v4.5.1`](https://github.com/cosmos/ibc-go/releases/tag/v4.5.1) ([#250](https://github.com/noble-assets/noble/pull/250))
- Bump Packet Forward Middleware to [`v4.1.1`](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.1) ([#250](https://github.com/noble-assets/noble/pull/250), [#258](https://github.com/noble-assets/noble/pull/258))

### FEATURES

- Include support for Coinbase's [Rosetta API](https://docs.cloud.coinbase.com/rosetta/docs/welcome). ([#215](https://github.com/noble-assets/noble/pull/215))

### IMPROVEMENTS

- Add `x/fiattokenfactory` interface changes required for CCTP. ([#241](https://github.com/noble-assets/noble/pull/241))

## v3.1.0

*Sep 15, 2023*

This is a minor release to the v3 Radon line.

In response to multiple IBC channels expiring on Noble's mainnet network, it was decided to expand the functionality of Noble's Maintenance Multisig to include IBC upgrade functionality (allowing expired clients to be changed).

### FEATURES

- Include support for IBC inside the ParamAuthority. ([#235](https://github.com/noble-assets/noble/pull/235))

### IMPROVEMENTS

- Align module path with Go's [naming convention](https://go.dev/doc/modules/version-numbers#major-version). ([#234](https://github.com/noble-assets/noble/pull/234))

---

## Previous Changes

This changelog has yet to be fully initialized. For previous versions please refer to the release notes for a summary of changes.

