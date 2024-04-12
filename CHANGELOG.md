# CHANGELOG

## v4.1.0

*Apr 12, 2024*

This is the first minor release to the v4 Argon release line, codenamed Fusion.

### DEPENDENCIES

- Bump PFM from Mandrake patch commit ([`455757b`](https://github.com/cosmos/ibc-apps/commit/455757bb5771c29cf2f83b59e37f6513e07c92be)) to release tag ([`v4.1.2`](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.2)). ([#346](https://github.com/noble-assets/noble/pull/346))
- Bump IBC to [`v4.6.0`](https://github.com/cosmos/ibc-go/releases/tag/v4.6.0) to resolve [ASA-2024-007](https://github.com/cosmos/ibc-go/security/advisories/GHSA-j496-crgh-34mx) security advisory. ([#346](https://github.com/noble-assets/noble/pull/346))
- Bump FiatTokenFactory to [`0a7385d`](https://github.com/circlefin/noble-fiattokenfactory/commit/0a7385d9a37744ced1e4d61eae10de2b117f482b) for various blocklist and paused check improvements. ([#346](https://github.com/noble-assets/noble/pull/346))

### FEATURES

- Introduce a new `x/forwarding` module for accounts that automatically forward cross-chain. ([#302](https://github.com/noble-assets/noble/pull/302))

### IMPROVEMENTS

- Align module path with Go's [naming convention](https://go.dev/doc/modules/version-numbers#major-version). ([#249](https://github.com/noble-assets/noble/pull/249))
- Remove `x/blockibc` middleware from codebase and switch to migrated version under [`circlefin/noble-fiattokenfactory`](https://github.com/circlefin/noble-fiattokenfactory). ([#346](https://github.com/noble-assets/noble/pull/346))

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

