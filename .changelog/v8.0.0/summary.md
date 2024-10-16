*Oct 17, 2024*

This is the Helium major release of Noble. It upgrades the Noble's core
dependencies, namely CometBFT (f.k.a. Tendermint), Cosmos SDK, and IBC to their
latest stable release [Eden]. In addition to this upgrade, it also replaces the
legacy [ParamAuthority] module with an in-house build [Authority] module. This
module allows the Noble Maintenance Multisig to enact governance gated actions
like chain upgrades and IBC client substitutions.

The following modules have specifically been upgraded to Cosmos SDK `v0.50.x`

- [FiatTokenFactory] — Circle's USD Coin
- [CCTP] — Circle's Cross Chain Transfer Protocol
- [Aura] — Ondo's US Dollar Yield Token
- [Halo] — Hashnote's US Yield Coin
- [Florin] — Monerium's EUR emoney
- [Forwarding] — Noble's Intents System

[aura]: https://github.com/ondoprotocol/usdy-noble
[authority]: https://github.com/noble-assets/authority
[cctp]: https://github.com/circlefin/noble-cctp
[eden]: https://medium.com/the-interchain-foundation/elevating-the-cosmos-sdk-eden-v0-50-20a554e16e43
[florin]: https://github.com/monerium/module-noble
[forwarding]: https://github.com/noble-assets/forwarding
[halo]: https://github.com/noble-assets/halo
[fiattokenfactory]: https://github.com/circlefin/noble-fiattokenfactory
[paramauthority]: https://github.com/strangelove-ventures/paramauthority
