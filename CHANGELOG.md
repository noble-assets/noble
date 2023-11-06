# CHANGELOG

## v4.0.0

*Nov 6, 2023*

This is the long awaited Argon major release of Noble. It introduces a new [`x/cctp`](https://github.com/circlefin/noble-cctp) module that implements Circle's [Cross Chain Transfer Protocol (CCTP)](https://www.circle.com/en/cross-chain-transfer-protocol), allowing native $USDC transfers between supported EVM networks and Noble (with many more networks to come). 

Along with the integration of the CCTP module, the following changes were made.

### BUG FIXES

- Fix simulation tests. ([#252](https://github.com/strangelove-ventures/noble/pull/252))
- Fix Ledger support for macOS Sonoma. ([#253](https://github.com/strangelove-ventures/noble/pull/253))

### DEPENDENCIES

- Bump IBC to [`v4.5.1`](https://github.com/cosmos/ibc-go/releases/tag/v4.5.1) ([#250](https://github.com/strangelove-ventures/noble/pull/250))
- Bump Packet Forward Middleware to [`v4.1.1`](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.1) ([#250](https://github.com/strangelove-ventures/noble/pull/250), [#258](https://github.com/strangelove-ventures/noble/pull/258))

### FEATURES

- Include support for Coinbase's [Rosetta API](https://docs.cloud.coinbase.com/rosetta/docs/welcome). ([#215](https://github.com/strangelove-ventures/noble/pull/215))

### IMPROVEMENTS

- Add `x/fiattokenfactory` interface changes required for CCTP. ([#241](https://github.com/strangelove-ventures/noble/pull/241))

## v3.1.0

*Sep 15, 2023*

This is a minor release to the v3 Radon line.

In response to multiple IBC channels expiring on Noble's mainnet network, it was decided to expand the functionality of Noble's Maintenance Multisig to include IBC upgrade functionality (allowing expired clients to be changed).

### FEATURES

- Include support for IBC inside the ParamAuthority. ([#235](https://github.com/strangelove-ventures/noble/pull/235))

### IMPROVEMENTS

- Align module path with Go's [naming convention](https://go.dev/doc/modules/version-numbers#major-version). ([#234](https://github.com/strangelove-ventures/noble/pull/234))

---

## Previous Changes

This changelog has yet to be fully initialized. For previous versions please refer to the release notes for a summary of changes.

