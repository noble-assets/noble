# CHANGELOG

## v3.1.1

*Oct 20, 2023*

This is a patch release to the v3.1 Radon line.

It includes a non-consensus breaking change to the `x/fiattokenfactory` module. This change allows the `Burn` and `Mint` methods to be called from other modules. Please note that the same permissions still apply to these methods.

### IMPROVEMENTS

- Add `x/fiattokenfactory` interface changes required for [CCTP](https://www.circle.com/en/cross-chain-transfer-protocol). ([#251](https://github.com/strangelove-ventures/noble/pull/251))

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

This changelog has yet to be fully initialized. For previous verions please refer to the release notes for a summary of changes.

