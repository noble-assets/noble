*Apr 15, 2024*

This is a minor release to the v4 Argon line, codenamed Fusion.

The main part of this release is the introduction of the `x/forwarding` module.
It allows users to create a new account type, that when receiving funds,
automatically sends those funds via IBC over a specified channel to a recipient
address. This allows for one-click transfers to any IBC enabled chain using
[Circle Mint][mint] or [CCTP].

Other notable changes include are documented below.

[cctp]: https://www.circle.com/en/cross-chain-transfer-protocol
[mint]: https://www.circle.com/en/circle-mint
