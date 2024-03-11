# CHANGELOG

## v4.0.3

*Mar 11, 2023*

This is the third patch release to the v4 Argon line.

### DEPENDENCIES

- Bump FiatTokenFactory to [`14edf83`](https://github.com/circlefin/noble-fiattokenfactory/commit/14edf83ee1c96055e2c17ea56ca9dd303d3c14f6) to enable `x/authz` support.
- Bump PFM to [`455757b`](https://github.com/cosmos/ibc-apps/commit/455757bb5771c29cf2f83b59e37f6513e07c92be) to resolve Mandrake disclosure.

### IMPROVEMENTS

- Switch to [migrated](https://github.com/circlefin/noble-fiattokenfactory) version of `x/fiattokenfactory` module.

## v4.0.2

*Nov 21, 2023*

This is the second patch release to the v4 Argon line.

### IMPROVEMENTS

- Implement a parameter query for the `x/tariff` module. ([#277](https://github.com/noble-assets/noble/pull/277))

## v4.0.1

*Nov 16, 2023*

This is the first patch release to the v4 Argon line.

### BUG FIXES

- Unregister `x/distribution` hooks to address consensus failure. ([#274](https://github.com/noble-assets/noble/pull/274))

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

---

## Other Releases

This changelog is specific to the v4 Argon release line. For other versions please refer to their release notes for a summary of changes.

