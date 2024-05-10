# CHANGELOG

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

## v4.1.1-rc.0

*Apr 16, 2024*

This is the release candidate for a consensus breaking patch release to the v4.1 Fusion line.

### BUG FIXES

- Remove custom ABCI logic inside `DeliverTx` that causes consensus failures. ([`8a4cf67`](https://github.com/noble-assets/noble/commit/8a4cf6768b3be88209c3fcdced146a0dfaf729e1))

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

## v4.1.0-rc.4

*Apr 14, 2024*

This is the fifth release candidate for a minor release to the v4 Argon line.

### BUG FIXES

- Improve checks around account existence when registering a new forwarding account. ([#323](https://github.com/noble-assets/noble/pull/323))
- Implement channel state checks registering and clearing forwarding accounts. ([#328](https://github.com/noble-assets/noble/pull/328))
- Pass a packet onto the next middleware if we fail to decode the receiver. ([#350](https://github.com/noble-assets/noble/pull/350))

### DEPENDENCIES

- Bump PFM from Mandrake patch commit ([`455757b`](https://github.com/cosmos/ibc-apps/commit/455757bb5771c29cf2f83b59e37f6513e07c92be)) to release tag ([`v4.1.2`](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.2)). ([#346](https://github.com/noble-assets/noble/pull/346))
- Switch to Noble's Cosmos SDK fork ([`v0.45.16-noble`](https://github.com/noble-assets/cosmos-sdk/releases/tag/v0.45.16-noble)), allowing `DeliverTx` to be extended. ([#346](https://github.com/noble-assets/noble/pull/346))
- Bump IBC to [`v4.6.0`](https://github.com/cosmos/ibc-go/releases/tag/v4.6.0) to resolve [ASA-2024-007](https://github.com/cosmos/ibc-go/security/advisories/GHSA-j496-crgh-34mx) security advisory. ([#346](https://github.com/noble-assets/noble/pull/346))
- Bump FiatTokenFactory to [`0a7385d`](https://github.com/circlefin/noble-fiattokenfactory/commit/0a7385d9a37744ced1e4d61eae10de2b117f482b) for various blocklist and paused check improvements. ([#346](https://github.com/noble-assets/noble/pull/346))

### FEATURES

- Allow forwarding accounts to be manually cleared by any user. ([#312](https://github.com/noble-assets/noble/pull/312))

### IMPROVEMENTS

- Switch to migrated `x/blockibc` under [`circlefin/noble-fiattokenfactory`](https://github.com/circlefin/noble-fiattokenfactory). ([#346](https://github.com/noble-assets/noble/pull/346))

## v4.1.0-rc.3

*Mar 11, 2024*

This is the fourth release candidate for a minor release to the v4 Argon line.

### BUG FIXES

- Correctly track `NumOfForwards` and `TotalForwarded` when forwarding. ([#310](https://github.com/noble-assets/noble/pull/310))

### DEPENDENCIES

- Bump FiatTokenFactory to [`14edf83`](https://github.com/circlefin/noble-fiattokenfactory/commit/14edf83ee1c96055e2c17ea56ca9dd303d3c14f6) to enable `x/authz` support.
- Bump PFM to [`455757b`](https://github.com/cosmos/ibc-apps/commit/455757bb5771c29cf2f83b59e37f6513e07c92be) to resolve Mandrake disclosure.

## v4.1.0-rc.2

*Feb 29, 2024*

This is the third release candidate for a minor release to the v4 Argon line.

### DEPRECATED

- Remove the new `x/stabletokenfactory` module for issuing [USDLR by Stable](https://withstable.com). ([#288](https://github.com/noble-assets/noble/pull/288))

### FEATURES

- Introduce a new `x/forwarding` module for accounts that automatically forward cross-chain. ([#302](https://github.com/noble-assets/noble/pull/302))

### IMPROVEMENTS

- Implement a parameter query for the `x/tariff` module. ([#277](https://github.com/noble-assets/noble/pull/277))
- Align module path with repository migration to `noble-assets` organization. ([#283](https://github.com/noble-assets/noble/pull/283))

## v4.1.0-rc.1

*Nov 16, 2023*

This is the second release candidate for a minor release to the v4 Argon line.

### BUG FIXES

- Unregister `x/distribution` hooks to address consensus failure. ([#275](https://github.com/noble-assets/noble/pull/275))

## v4.1.0-rc.0

*Nov 15, 2023*

This is the first release candidate for a minor release to the v4 Argon line.

### DEPENDENCIES

- Bump CCTP from [`dc81b3e`](https://github.com/circlefin/noble-cctp/commit/dc81b3e0d566d195c869a213519fcecd38b020a5) to [`86f425e`](https://github.com/circlefin/noble-cctp/commit/86f425e6fac94ff57865dd11b42c95de303e0d96) ([#259](https://github.com/noble-assets/noble/pull/259))

### FEATURES

- Introduce a new `x/stabletokenfactory` module for issuing [USDLR by Stable](https://withstable.com). ([#269](https://github.com/noble-assets/noble/pull/269))

### IMPROVEMENTS

- Align module path with Go's [naming convention](https://go.dev/doc/modules/version-numbers#major-version). ([#249](https://github.com/noble-assets/noble/pull/249))
- Remove `x/fiattokenfactory` module from codebase and switch to migrated version under [`circlefin/noble-fiattokenfactory`](https://github.com/circlefin/noble-fiattokenfactory) ([#259](https://github.com/noble-assets/noble/pull/259))
- Add multiple fee denom support to the `x/tariff` module. ([#269](https://github.com/noble-assets/noble/pull/269))

---

## Other Releases

This changelog is specific to the v4.1 Argon release line. For other versions please refer to their release notes for a summary of changes.

