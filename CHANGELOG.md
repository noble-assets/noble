# CHANGELOG

## v6.0.0

*Aug 26, 2024*

This is the Xenon major release of Noble. It introduces a new `x/halo` module
that enables the native issuance of [Hahsnote's US Yield Coin (**USYC**)][usyc]
asset. USYC is an on-chain representation of Hashnote's Short Duration Yield
Fund, primarily investing in U.S. Treasury Bills and engaging in reverse repo
activities.

Other notable changes are documented below.

[usyc]: https://usyc.hashnote.com

### IMPROVEMENTS

- Update module path for v6 release line. ([#389](https://github.com/noble-assets/noble/pull/389))

## v5.0.0

*Jul 5, 2024*

This is the Krypton major release of Noble. It introduces a new `x/aura` module
that enables the native issuance of [Ondo's US Dollar Yield (**USDY**)][usdy]
asset. USDY is a tokenized note secured by short-term US Treasuries and bank
demand deposits.

Other notable changes are documented below.

[usdy]: https://ondo.finance/usdy

### DEPENDENCIES

- Switch to Noble's Cosmos SDK fork ([`v0.45.16-send-restrictions`](https://github.com/noble-assets/cosmos-sdk/releases/tag/v0.45.16-send-restrictions)) that supports send restrictions. ([#385](https://github.com/noble-assets/noble/pull/385))

### FEATURES

- Update the default `commit_timeout` to `500ms` to improve block time. ([#380](https://github.com/noble-assets/noble/pull/380))

### IMPROVEMENTS

- Update module path for v5 release line. ([#271](https://github.com/noble-assets/noble/pull/271))

## v4.1.3

*May 10, 2024*

This is a consensus breaking patch release to the v4.1 Fusion line.

### DEPENDENCIES

- Bump CCTP to [`253cf7e`](https://github.com/circlefin/noble-cctp/commit/253cf7eb943669e283b4dcb25f83c7096080e67a) ([#363](https://github.com/noble-assets/noble/pull/363))

## v4.1.2

*May 2, 2024*

This is a consensus breaking patch release to the v4.1 Fusion line.

### DEPENDENCIES

- Bump `x/forwarding` module to [`v1.1.0`](https://github.com/noble-assets/forwarding/releases/tag/v1.1.0) ([#357](https://github.com/noble-assets/noble/pull/357))
- Bump FiatTokenFactory to [`738932c`](https://github.com/circlefin/noble-fiattokenfactory/commit/738932cb316d06f587c49dfb11a50515cce657d9) ([#359](https://github.com/noble-assets/noble/pull/359))
- Bump CCTP to [`69ee090`](https://github.com/circlefin/noble-cctp/commit/69ee090808c05987c504b383939e71ad491594e7) ([#359](https://github.com/noble-assets/noble/pull/359))

### IMPROVEMENTS

- Switch to [migrated](https://github.com/noble-assets/forwarding) version of `x/forwarding` module. ([#357](https://github.com/noble-assets/noble/pull/357))

## v4.1.1

*Apr 16, 2024*

This is a consensus breaking patch release to the v4.1 Fusion line.

### BUG FIXES

- Remove custom ABCI logic inside `DeliverTx` that causes consensus failures. ([#353](https://github.com/noble-assets/noble/pull/353))

## v4.1.0

*Apr 15, 2024*

This is a minor release to the v4 Argon line, codenamed Fusion.

The main part of this release is the introduction of the `x/forwarding` module.
It allows users to create a new account type, where the receipt of funds into
that account triggers an automatic IBC transfer over a specified channel to a
recipient address. This allows for one-click transfers to any IBC-enabled chain,
and can be used in tandem with, for example, the receipt of funds from a
[Circle Mint][mint] account or via [CCTP].

Other notable changes include are documented below.

[cctp]: https://www.circle.com/en/cross-chain-transfer-protocol
[mint]: https://www.circle.com/en/circle-mint

### DEPENDENCIES

- Switch to Noble's Cosmos SDK fork ([`v0.45.16-noble`](https://github.com/noble-assets/cosmos-sdk/releases/tag/v0.45.16-noble)), allowing `DeliverTx` to be extended.
- Bump PFM from Mandrake patch commit ([`455757b`](https://github.com/cosmos/ibc-apps/commit/455757bb5771c29cf2f83b59e37f6513e07c92be)) to release tag ([`v4.1.2`](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.2)).
- Bump IBC to [`v4.6.0`](https://github.com/cosmos/ibc-go/releases/tag/v4.6.0) to resolve [ASA-2024-007](https://github.com/cosmos/ibc-go/security/advisories/GHSA-j496-crgh-34mx) security advisory.
- Bump FiatTokenFactory to [`0a7385d`](https://github.com/circlefin/noble-fiattokenfactory/commit/0a7385d9a37744ced1e4d61eae10de2b117f482b) for various blocklist and paused check improvements.

### IMPROVEMENTS

- Align module path with Go's [naming convention](https://go.dev/doc/modules/version-numbers#major-version). ([#249](https://github.com/noble-assets/noble/pull/249), [#283](https://github.com/noble-assets/noble/pull/283))
- Switch to migrated `x/blockibc` under [`circlefin/noble-fiattokenfactory`](https://github.com/circlefin/noble-fiattokenfactory). ([#346](https://github.com/noble-assets/noble/pull/346))

## v4.0.3

*Mar 11, 2024*

This is a consensus breaking patch release to the v4 Argon line.

### DEPENDENCIES

- Bump FiatTokenFactory to [`14edf83`](https://github.com/circlefin/noble-fiattokenfactory/commit/14edf83ee1c96055e2c17ea56ca9dd303d3c14f6) to enable `x/authz` support.
- Bump PFM to [`455757b`](https://github.com/cosmos/ibc-apps/commit/455757bb5771c29cf2f83b59e37f6513e07c92be) to resolve Mandrake disclosure.

### IMPROVEMENTS

- Switch to [migrated](https://github.com/circlefin/noble-fiattokenfactory) version of `x/fiattokenfactory` module.

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

