![Noble banner](https://raw.githubusercontent.com/noble-assets/networks/main/Twitter_Banner.png)
# ✨✨ Noble ✨✨

[Noble](https://noble.xyz/) is a Cosmos [application-specific blockchain](https://docs.cosmos.network/v0.50/learn/intro/why-app-specific) purpose-built for native asset issuance. Noble brings the efficiency and interoperability of native assets to the wider Cosmos ecosystem, supporting [Circle's USDC](https://www.circle.com/en/usdc), [Ondo's USDY](https://ondo.finance/), [Monerium's EURe](https://monerium.com/). Noble’s vision is to be the world’s premier issuance hub for digital assets that connect to other blockchains seamlessly. Noble leverages the Cosmos-SDK – a flexible toolkit that allows developers to leverage existing modules, and to seamlessly integrate custom modules that add virtually unlimited functionality for asset issuers onthe Noble blockchain.

[![twitter](https://img.shields.io/badge/twitter-@noble_xyz-white?labelColor=0f1331&style=flat)](https://twitter.com/noble_xyz/) [![telegram](https://img.shields.io/badge/telegram-noble-white?labelColor=0f1331&style=flat)](https://t.me/+5mCog0PAWCRmOWYx) [![website](https://img.shields.io/badge/website-noble.xyz-white?labelColor=0f1331&style=flat)](https://noble.xyz/)

## Contents
1. [Noble Chain Design](#noble-chain-design)
2. [Assets on Noble](#assets-on-noble)
3. [Protocol Features](#protocol-features)
4. [Noble on the Interchain](#noble-on-the-interchain)
5. [Noble Upgrades and Governance](#noble-upgrades--governance)
6. [For Developers](#for-developers)
7. [For Node Operators & Validators](#for-node-operators--validators)
8. [Contributing](#contributing)

## Noble Chain Design

The Noble protocol runs on Proof-of-Authority consensus with a trusted validator set. The validator set is permissioned by equal shares of staking tokens placed in vesting accounts. The tokens have no value. Economic security is derived from fees captured by the chain as issued  assets on a block-by-block basis. If a double sign is detected by the chain, the validator address will be “tombstoned”, meaning that their tokens and the address will no longer be usable for validation and resulting in the loss of all future fee revenue. The Proof-of-Authority model provides real economic cost to faulty validator behavior and thus provides economic security to the network that can be computed in real time based on past and projected fees.

Noble intends to monitor developments in shared security across the blockchain ecosystem to ensure the optimality of the security model.

## Assets on Noble

Every asset issued on the Noble chain is implemented as its own Cosmos SDK module.

| Issuer                                   | Asset | Module Name            | 
| ---------------------------------------- | ----- | ---------------------- | 
| [Circle](https://www.circle.com/en/usdc) | USDC  | [noble-fiattokenfactory](https://github.com/circlefin/noble-fiattokenfactory) |
| [Ondo](https://ondo.finance/)            | USDY  | [aura](https://github.com/ondoprotocol/usdy-noble)                       | 
| [Monerium](https://monerium.com/)        | EURe  | [florin](https://github.com/monerium/module-noble)                     |
| [Hashnote](https://www.hashnote.com/)    | USYC  | [halo](https://github.com/noble-assets/halo)                             |

## Protocol features

Along with asset issuance, the Noble protocol also implements the following modules:

| Module name                                                              | Features |
| ------------------------------------------------------------------------ | -------- |
| [noble-cctp](https://github.com/circlefin/noble-cctp)                    | Allows native transfer of Circle's USDC across ecosystems through their [Cross Chain Transfer Protocol](https://www.circle.com/en/cross-chain-transfer-protocol) |
| [forwarding](https://github.com/noble-assets/forwarding)                 | Allows a custom account type on Noble to auto forward any tokens it receives to an address on a different chain connected via IBC |
| [globalfee](./x/globalfee)                                               | Allows the protocol to control the transaction fees and bypass gas fees for specific sdk.Msg types |
| [tariff](./x/tariff)                                                     | Adds a source of revenue for the protocol as well as its contributors by collecting a percentage of share of token transfers via Noble | 
| [paramauthority](https://github.com/strangelove-ventures/paramauthority) | Allows a configured address to perform the governance operations. [More](#noble-upgrades--governance) | 

## Noble on the Interchain

App-chains are able to permissionlessly connect to Noble [via IBC](https://medium.com/the-interchain-foundation/eli5-what-is-ibc-def44d7b5b4c), a universal interoperability protocol that allows two different blockchains to communicate with one another, garaunteeing reliable, ordered, and authenticated communication.

#### How to integrate with Noble? 

To establish a connection to Noble (or any 2 IBC compatible chains), anybody can set up a relayer on an existing path. This [tutorial](https://github.com/cosmos/relayer/blob/main/docs/create-path-across-chain.md) gives an overview of how to create a channel, client, and connection and to start relaying IBC packets.

## Noble Upgrades & Governance

The Noble blockchain (“Noble”) will house a module, referred to as the Noble Maintenance Multisig (or “NMM”), that has the following three functionalities:

- Upgrade authority
- Parameter Changes
- IBC maintenance

**Upgrade Authority** 

A standard multisignature application (the “Noble Multisig”) will form the Noble Maintenance Multisig (“NMM”) which will be able to initiate a chain upgrade when a breaking protocol change requires upgrading the chain binary. The NMM will use an on-chain procedure to instruct the validators to halt the protocol running the Noble chain and which will restart the blockchain with a new binary. Importantly, validators have discretionary authority to follow those instructions. If 1/3 + 1 of the voting power of the Noble chain refuses to follow these upgrade instructions, the chain halts.

**Parameter Changes**

The NMM has the ability to initiate parameter changes. This authority is distinct from the upgrade functionality as it would not require the Noble validators to run a new chain binary. A parameter change would automatically be introduced into the state machine of the Noble chain.

**IBC Connection Maintenance:**

The NMM has the ability to both re-establish IBC connections that have expired as well as sever IBC connections. Any such action related to an IBC connection is achieved automatically upon execution by the Noble Multisig. This documentation further outlines situations where IBC maintenance would be required.

**Configuration of NMM** 

The NMM is a 5/7 multisig for passing a Noble Multisig proposal to fulfill the three key functions discussed above. The configuration of the Noble Multisig will include the Noble corporate entity and core validator partners. A key objective of the NMM is to achieve a reasonable level of geographic diversity among Noble Multisig members for resilience. The proposed geographies are noted as follows:

- Noble (the organization): Canada 
- Iqlusion: USA 
- Strangelove: USA 
- Chorus One: Switzerland
- Binary Holdings: Switzerland 
- Cosmostation: Korea  
- Luganodes: Switzerland 

Stakeholders on the NMM will be able to be rotated in and out of the mulitisg over time. 

## For Developers 
 
### Setup Noble locally

The following software should be installed on the target system:

1. Go Programming Language (https://go.dev)
2. Git  (https://git-scm.com)
3. GNU Make (https://www.gnu.org/software/make)

```
git clone https://github.com/noble-assets/noble
cd noble
make install
```

### Local Running

To quickly spin up a standalone noble chain and setup all privileged accounts, run the [play.sh](play.sh) bash script

```
sh play.sh
```

## For Node Operators & Validators

### Running Nodes

- See [`noble-networks`](https://github.com/noble-assets/networks/tree/main) for all info around validating for mainnet and testent

## Contributing

This repository is not looking for external contributors.