<div align="center">
  <picture>
    <img src="https://github.com/user-attachments/assets/5eafb881-ce88-4b16-b61e-019da8327def" />
  </picture>
  <br><br>
  <a href="https://github.com/noble-assets/noble/releases/latest"><img alt="Latest Release" src="https://img.shields.io/github/v/release/noble-assets/noble?style=for-the-badge&label=Latest%20Release&color=BAC3FF"></a>
  <a href="https://mintscan.io/noble/block"><img alt="Block Height" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Frpc.noble.xyz%2Fblock&query=result.block.header.height&style=for-the-badge&label=Block%20Height&color=BAC3FF"></a>
  <br><br>
  <a href="https://noble.xyz">✨ noble.xyz ✨</a>
  <br>
  <a href="https://dollar.noble.xyz">✨ dollar.noble.xyz ✨</a>
  <br>
  <a href="https://express.noble.xyz">✨ express.noble.xyz ✨</a>
</div>
<br>

Noble is an [application-specific blockchain](https://docs.cosmos.network/main/learn/intro/why-app-specific) built on the Cosmos SDK, purpose-built for asset issuance with a focus on stablecoins and real-world assets (RWAs). As an IBC-enabled chain, Noble ensures seamless interoperability across the Cosmos ecosystem, enabling fast and secure transactions.

In addition to supporting various RWAs, Noble offers its own native real-world asset, Noble Dollar (USDN)—a yield-bearing stablecoin that gives developers and end users control over the underlying yield. Additionally, Noble has implemented Circle's [Cross-Chain Transfer Protocol (CCTP)](https://www.circle.com/cross-chain-transfer-protocol) to facilitate transfers of USDC across multiple blockchain networks.

You can learn more about all the assets we offer [here](https://www.noble.xyz/#assets)!

## Documentation

For all documentation outside of installation, please visit our [official documentation](https://docs.noble.xyz).

## Installation

Install from source:

[Golang](https://go.dev/) is required.

```sh
git clone https://github.com/noble-assets/noble.git
cd noble
git checkout <TAG>
make install
```

Noble is also available via:

- [Releases](https://github.com/noble-assets/noble/releases)
- [Docker](https://github.com/noble-assets/noble/pkgs/container/noble)

## Local Net Quickstart

Looking to spin up a stand alone local Noble chain? Leverage our quickstart guide and scripts [here](./local_net/)!

## Contributing

We welcome contributions! If you find a bug or have feedback, open an issue. Pull requests for bug fixes are appreciated.
For major changes, please open an issue first to discuss your proposal.

Note: We do not accept contributions for grammar or spelling corrections.
