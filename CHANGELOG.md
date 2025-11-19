# CHANGELOG

## v11.1.0

*Nov 19, 2025*

This is a minor release to the v11 Flux line, codenamed Pulse.

### DEPENDENCIES

- Bump Orbiter to [`v2.0.0`](https://github.com/noble-assets/orbiter/releases/tag/v2.0.0) to support a fixed amount fee and improve stats. ([#609](https://github.com/noble-assets/noble/pull/609), [#611](https://github.com/noble-assets/noble/pull/611))

## v11.0.1

*Oct 21, 2025*

This is a non-consensus breaking patch to the v11 Flux release line.

No core code changes are made in this patch.

### IMPROVEMENTS

- Ensure binaries included in releases are statically linked. ([#605](https://github.com/noble-assets/noble/pull/605))

## v11.0.0

*Oct 21, 2025*

This is the Flux major release of Noble. It introduces the [Orbiter](https://github.com/noble-assets/orbiter) module, that is a new system for cross-chain interoperability.

This and other notable changes are documented below.

### DEPENDENCIES

- Bump Dollar to [`v2.2.0`](https://github.com/noble-assets/dollar/releases/tag/v2.2.0) to migrate Points Season Two configuration values to state. ([#573](https://github.com/noble-assets/noble/pull/573))

### FEATURES

- Integrate Orbiter module, that introduces a new system for cross-chain interoperability. ([#576](https://github.com/noble-assets/noble/pull/576))
- Enable permissionless Hyperlane collateral token creation, for all assets except $USDN. ([#587](https://github.com/noble-assets/noble/pull/587))

### IMPROVEMENTS

- Remove Heighliner with a Dockerfile in E2E and build environments. ([#557](https://github.com/noble-assets/noble/pull/557))
- Update module path for v11 release line. ([#564](https://github.com/noble-assets/noble/pull/564))
- Claim `x/distribution` module funds via upgrade handler that were stuck after its removal. ([#571](https://github.com/noble-assets/noble/pull/571))
- Recover expired Shido IBC light client, `07-tendermint-106` → `07-tendermint-186` ([#597](https://github.com/noble-assets/noble/pull/597))
- Recover expired Router IBC light client, `07-tendermint-136` → `07-tendermint-192` ([#603](https://github.com/noble-assets/noble/pull/603))

## v10.1.2

*Oct 14, 2025*

This is a consensus breaking patch to the v10.1 Ember release line.

### DEPENDENCIES

- Bump Forwarding to [`v2.0.3`](https://github.com/noble-assets/forwarding/releases/tag/v2.0.3) to remove `x/bank` `GetAllBalances` usage. ([#600](https://github.com/noble-assets/noble/pull/600))
- Bump CometBFT to [`v0.38.19`](https://github.com/cometbft/cometbft/releases/tag/v0.38.19) to resolve [ASA-2025-003](https://github.com/cometbft/cometbft/security/advisories/GHSA-hrhf-2vcr-ghch) security advisory. ([#600](https://github.com/noble-assets/noble/pull/600))

## v10.1.1

*Aug 25, 2025*

This is a non-consensus breaking patch to the v10.1 Ember release line.

This upgrade is only relevant for validators.

### DEPENDENCIES

- Bump GlobalFee to [`v1.0.1`](https://github.com/noble-assets/globalfee/releases/tag/v1.0.1) to harden mempool checks of bypass messages. ([#582](https://github.com/noble-assets/noble/pull/582))

## v10.1.0

*Jul 30, 2025*

This is a minor release to the v10 Stratum line, codenamed Ember.

### DEPENDENCIES

- Bump Dollar to [`v2.1.0`](https://github.com/noble-assets/dollar/releases/tag/v2.1.0) to introduce logic that handles the end of Vaults Season One. ([#570](https://github.com/noble-assets/noble/pull/570))

## v10.0.1

*Jul 7, 2025*

This is a patch to the v10 Stratum release line.

If the Dollar module has no yield recipients enabled, this is non-consensus breaking.

### DEPENDENCIES

- Bump Dollar to [`v2.0.1`](https://github.com/noble-assets/dollar/releases/tag/v2.0.1) to gracefully handle transfer errors. ([#567](https://github.com/noble-assets/noble/pull/567))

## v10.0.0

*Jun 6, 2025*

This is the Stratum major release of Noble. It introduces [composable yield](https://www.noble.xyz/blog/composable-yield-a-new-paradigm-for-stablecoins) for the Noble Dollar (USDN), along with integration of the [Hyperlane](https://hyperlane.xyz/) protocol.

This and other notable changes are documented below.

### BUG FIXES

- Ensure transaction injections from Jester are non-empty. ([#515](https://github.com/noble-assets/noble/pull/515))

### DEPENDENCIES

- Bump supported Golang version to `v1.24` ([#530](https://github.com/noble-assets/noble/pull/530))

### FEATURES

- Integrate the Hyperlane Core module, that enables messaging via the Hyperlane protocol. ([#519](https://github.com/noble-assets/noble/pull/519))
- Upgrade the Dollar module to enable $USDN yield distribution across specific IBC channels and Hyperlane routes. ([#526](https://github.com/noble-assets/noble/pull/526))
- Integrate the Hyperlane Warp module, that enables token transfers via the Hyperlane protocol. ([#527](https://github.com/noble-assets/noble/pull/527))
- Introduce a new ante handler that permissions Hyperlane actions. ([#528](https://github.com/noble-assets/noble/pull/528))
- Integrate the IBC Rate Limit module, that enables more granular control over IBC token transfers. ([#541](https://github.com/noble-assets/noble/pull/541))

### IMPROVEMENTS

- Update module path for v10 release line. ([#516](https://github.com/noble-assets/noble/pull/516))

## v9.0.4

*May 8, 2025*

This is a non-consensus breaking patch to the v9 Argentum release line.

### DEPENDENCIES

- Bump Dollar to [`v1.0.2`](https://github.com/noble-assets/dollar/releases/tag/v1.0.2) to improve `PendingRewards` query. ([#549](https://github.com/noble-assets/noble/pull/549))
- Bump Swap to [`v1.0.2`](https://github.com/noble-assets/swap/releases/tag/v1.0.2) to improve `SimulateSwap` query. ([#549](https://github.com/noble-assets/noble/pull/549))

## v9.0.3

*Mar 25, 2025*

This is a consensus breaking patch to the v9 Argentum release line.

### DEPENDENCIES

- Bump Swap to [`v1.0.1`](https://github.com/noble-assets/swap/releases/tag/v1.0.1) ([#522](https://github.com/noble-assets/noble/pull/522))

## v9.0.2

*Mar 12, 2025*

This is a non-consensus breaking patch to the v9 Argentum release line.

### DEPENDENCIES

- Bump IBC to [`v8.7.0`](https://github.com/cosmos/ibc-go/releases/tag/v8.7.0) to resolve [ISA-2025-001](https://github.com/cosmos/ibc-go/security/advisories/GHSA-4wf3-5qj9-368v) security advisory. ([#513](https://github.com/noble-assets/noble/pull/513))

## v9.0.1

*Mar 4, 2025*

This is a non-consensus breaking patch to the v9 Argentum release line.

### DEPENDENCIES

- Bump Authority to [`v1.0.3`](https://github.com/noble-assets/authority/releases/tag/v1.0.3) to correctly implement codec interface. ([#509](https://github.com/noble-assets/noble/pull/509))
- Bump Dollar to [`v1.0.1`](https://github.com/noble-assets/dollar/releases/tag/v1.0.1) to correct recipient address in event. ([#511](https://github.com/noble-assets/noble/pull/511))

## v9.0.0

*Feb 28, 2025*

This is the Argentum major release of Noble. It introduces various new modules
that enable the issuance and use-cases of the Noble Dollar (USDN), Noble's
yield bearing stablecoin. USDN is fully collateralized by U.S. Treasury bills
via the M^0 protocol.

This and other notable changes are documented below.

### BUG FIXES

- Update the capabilities of previously created ICA channels from the ICA Controller module back to the ICA Host module. ([#432](https://github.com/noble-assets/noble/pull/432))

### DEPENDENCIES

- Bump FiatTokenFactory to remove the limit check when decoding addresses. ([#455](https://github.com/noble-assets/noble/pull/455))
- Bump `cosmossdk.io/client/v2` to support returning maps inside AutoCLI queries. ([#464](https://github.com/noble-assets/noble/pull/464))
- Bump Authority to [`v1.0.2`](https://github.com/noble-assets/authority/releases/tag/v1.0.2) to include a new helper CLI command. ([#480](https://github.com/noble-assets/noble/pull/480))
- Bump Forwarding to [`v2.0.1`](https://github.com/noble-assets/forwarding/releases/tag/v2.0.1) to check recipient length and harden validation when registering accounts. ([#481](https://github.com/noble-assets/noble/pull/481))
- Bump Packet Forward Middleware to [`v8.2.0`](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv8.2.0) to resolve [GHSA-6fgm-x6ff-w78f](https://github.com/cosmos/ibc-apps/security/advisories/GHSA-6fgm-x6ff-w78f) security advisory. ([#488](https://github.com/noble-assets/noble/pull/488), [#506](https://github.com/noble-assets/noble/pull/506))
- Bump Cosmos SDK to [`v0.50.12`](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.12) ([#495](https://github.com/noble-assets/noble/pull/495))
- Bump IBC to [`v8.6.1`](https://github.com/cosmos/ibc-go/releases/v8.6.1) to resolve [ASA-2025-004](https://github.com/cosmos/ibc-go/security/advisories/GHSA-jg6f-48ff-5xrw) security advisory. ([#506](https://github.com/noble-assets/noble/pull/506))

### FEATURES

- Integrate our custom Wormhole module, that enables Wormhole messaging on Noble via IBC. ([#444](https://github.com/noble-assets/noble/pull/444))
- Integrate our custom Dollar module, that enables the issuance of Noble's stablecoin $USDN. ([#448](https://github.com/noble-assets/noble/pull/448))
- Integrate our custom Swap module, that enables the exchange of tokens issued on Noble. ([#449](https://github.com/noble-assets/noble/pull/449))
- Integrate our custom Jester sidecar, that enables the automatic relaying of $USDN transfers to Noble. ([#463](https://github.com/noble-assets/noble/pull/463))
- Enable Swagger documentation in API endpoint. ([#470](https://github.com/noble-assets/noble/pull/470))
- Enable functionality for in-place forking a synced testnet or mainnet node. ([#487](https://github.com/noble-assets/noble/pull/487))

### IMPROVEMENTS

- Update module path for v9 release line. ([#443](https://github.com/noble-assets/noble/pull/443))

## v8.0.5

*Feb 3, 2025*

This is a non-consensus breaking patch to the v8 Helium release line.

### DEPENDENCIES

- Bump CometBFT to [`v0.38.17`](https://github.com/cometbft/cometbft/releases/v0.38.17) to resolve [ASA-2025-001](https://github.com/cometbft/cometbft/security/advisories/GHSA-22qq-3xwm-r5x4) and [ASA-2025-002](https://github.com/cometbft/cometbft/security/advisories/GHSA-r3r4-g7hq-pq4f) security advisories. ([#466](https://github.com/noble-assets/noble/pull/466))

## v8.0.4

*Dec 16, 2024*

This is a non-consensus breaking patch to the v8 Helium release line.

### DEPENDENCIES

- Update `x/authority` to include helper CLI commands. ([#440](https://github.com/noble-assets/noble/pull/440))
- Bump Cosmos SDK to [`v0.50.11`](https://github.com/cosmos/cosmos-sdk/releases/v0.50.11) to resolve [ABS-0043/ABS-0044](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-8wcc-m6j2-qxvm) security advisory. ([#441](https://github.com/noble-assets/noble/pull/441))

## v8.0.3

*Nov 20, 2024*

This is a non-consensus breaking patch to the v8 Helium release line.

### DEPENDENCIES

- Update `x/halo` to latest non-consensus breaking patch. ([#431](https://github.com/noble-assets/noble/pull/431))
- Bump `cosmossdk.io/math` to [`v1.4.0`](https://github.com/cosmos/cosmos-sdk/releases/tag/math%2Fv1.4.0) to resolve [ASA-2024-010](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-7225-m954-23v7) security advisory. ([#433](https://github.com/noble-assets/noble/pull/433))

## v8.0.2

*Nov 13, 2024*

This is a consensus breaking patch to the v8 Helium release line.

It addresses the following consensus failure when upgrading Noble's mainnet to the `v8.0.1` release.

### BUG FIXES

- Due to IBC-Go v8 not supporting App Wiring, the Noble Core Team has to manually initialize all IBC modules and keepers. The Forwarding module receives multiple IBC keepers, which have to be manually set once wiring is complete. ([#429](https://github.com/noble-assets/noble/pull/429))

## v8.0.1

*Nov 13, 2024*

This is a consensus breaking patch to the v8 Helium release line.

It addresses the following consensus failure when upgrading Noble's mainnet to the `v8.0.0` release.

### BUG FIXES

- In the v8 Helium upgrade handler, the Noble Core Team wanted to align a missconfiguration in the initial genesis file that resulted in 18 surplus $STAKE existing, bringing the total supply to 1,000,000,018. The migration plan involved burning the surplus 18 tokens via the Uupgrade module, however, the module account was never initialized and permissioned. ([#428](https://github.com/noble-assets/noble/pull/428))

## v8.0.0

*Nov 11, 2024*

This is the Helium major release of Noble. It upgrades Noble itself and all of it's core modules to the latest stable release of the Cosmos SDK, `v0.50.x` a.k.a. Eden. Additional module changes have been documented below:

#### FiatTokenFactory

The BlockIBC logic was improved to support both Bech32 and Bech32m for IBC recipient addresses.

#### Florin

The module was updated to accept a user's public key when verifying signatures, instead of relying on on-chain data.

#### Forwarding

The module was updated to include a fallback address and a list of allowed denominations to forward.

## v7.0.0

*Sep 13, 2024*

This is the Numus major release of Noble. It introduces a new `x/florin` module
that enables the native issuance of [Monerium's EUR emoney (**EURe**)][eure]
asset. EURe is issued by Monerium EMI, a regulated entity, licensed in the EEA.
E-money is recognized as a digital alternative to cash, 1:1 backed in
high-quality liquid assets and is unconditionally redeemable on demand.

Other notable changes are documented below.

[eure]: https://monerium.com

### BUG FIXES

- Update `x/halo` to correctly check recipient role when trading to fiat. ([#405](https://github.com/noble-assets/noble/pull/405))

### IMPROVEMENTS

- Update module path for v7 release line. ([#399](https://github.com/noble-assets/noble/pull/399))

## v6.0.0

*Aug 27, 2024*

This is the Xenon major release of Noble. It introduces a new `x/halo` module
that enables the native issuance of [Hashnote's US Yield Coin (**USYC**)][usyc]
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
[Circle Mint][mint] account or via [CCTP][cctp-docs].

Other notable changes include are documented below.

[cctp-docs]: https://www.circle.com/en/cross-chain-transfer-protocol
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

