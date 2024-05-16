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
