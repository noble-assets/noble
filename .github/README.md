![](banner.png)

# ✨✨ Noble ✨✨

[Noble](https://noble.xyz) is a Cosmos [application-specific blockchain](https://docs.cosmos.network/v0.50/learn/intro/why-app-specific) purpose-built for native asset issuance. Noble brings the efficiency and interoperability of native assets to the wider Cosmos ecosystem, supporting [Circle's USDC](https://www.circle.com/en/usdc), [Ondo's USDY](https://ondo.finance/usdy), [Monerium's EURe](https://monerium.com), and [Hashnote's USYC](https://usyc.hashnote.com). Noble's vision is to be the world's premier issuance hub for digital assets that connect to other blockchains seamlessly. Noble leverages the Cosmos SDK – a flexible toolkit that allows developers to leverage existing modules, and to seamlessly integrate custom modules that add virtually unlimited functionality for asset issuers on the Noble blockchain.

[![twitter](https://img.shields.io/badge/twitter-@noble__xyz-white?labelColor=black)](https://twitter.com/noble_xyz) [![website](https://img.shields.io/badge/website-noble.xyz-white?labelColor=black)](https://noble.xyz)

## Contents
1. [Noble Chain Design](#noble-chain-design)
2. [Assets on Noble](#assets-on-noble)
3. [Protocol Features](#protocol-features)
4. [Noble on the Interchain](#noble-on-the-interchain)
5. [Noble Upgrades and Governance](#noble-upgrades--governance)
6. [For Developers](#for-developers)
7. [For Node Operators & Validators](#for-node-operators--validators)

## Noble Chain Design

The Noble chain runs on Proof-of-Authority consensus with a trusted validator set. The validator set is permissioned by equal shares of staking tokens placed in permanently locked vesting accounts. The tokens have no value. Economic security is derived from fees captured by the chain on a block-by-block basis. If a double sign is detected by the chain, the validator address will be "tombstoned", meaning that their token and address will no longer be usable for validation and resulting in the loss of all future fee revenue. The Proof-of-Authority model provides real economic cost to faulty validator behavior and thus provides economic security to the network that can be computed in real time based on past and projected fees.

The Noble Core Team intends to monitor developments in shared security across various ecosystems to ensure the optimality of our security model.

## Assets on Noble

Every asset issued on the Noble chain is implemented as its own dedicated Cosmos SDK module.

| Issuer                               | Asset | Module Name                                                               | 
|--------------------------------------|-------|---------------------------------------------------------------------------| 
| [Circle](https://www.circle.com)     | USDC  | [`fiattokenfactory`](https://github.com/circlefin/noble-fiattokenfactory) |
| [Ondo](https://ondo.finance)         | USDY  | [`aura`](https://github.com/ondoprotocol/usdy-noble)                      | 
| [Monerium](https://monerium.com)     | EURe  | [`florin`](https://github.com/monerium/module-noble)                      |
| [Hashnote](https://www.hashnote.com) | USYC  | [`halo`](https://github.com/noble-assets/halo)                            |

## Protocol Features

Along with asset issuance, the Noble chain also implements the following modules:

| Module Name                                                | Description                                                                                                                                                               |
|------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [`cctp`](https://github.com/circlefin/noble-cctp)          | Allows native transfers of Circle's stablecoins across ecosystems through their [Cross Chain Transfer Protocol](https://www.circle.com/en/cross-chain-transfer-protocol). |
| [`forwarding`](https://github.com/noble-assets/forwarding) | Allows specific accounts to be created to allow users to interact with IBC via a native bank transfer.                                                                    |
| [`globalfee`](https://github.com/noble-assets/globalfee)   | Allows Noble's governance to control transaction fees and certain transactions that can bypass this.                                                                      |
| [`authority`](https://github.com/noble-assets/authority)   | Allows the Noble Maintenance Multisig to perform governance actions on specific modules. [More](#noble-upgrades--governance).                                             | 

## Noble on the Interchain

App-chains are able to permissionlessly connect to Noble [via IBC](https://medium.com/the-interchain-foundation/eli5-what-is-ibc-def44d7b5b4c), a universal interoperability protocol that allows two different blockchains to communicate with one another, guaranteeing reliable, ordered, and authenticated communication.

### How to integrate with Noble? 

To establish a connection to Noble (or any 2 IBC compatible chains), anybody can set up a relayer on an existing path. This [tutorial](https://github.com/cosmos/relayer/blob/main/docs/create-path-across-chain.md) gives an overview of how to create a channel, client, and connection and to start relaying IBC packets.

## Noble Upgrades & Governance

The Noble chain is governed by the Noble Maintenance Multisig (or "NMM"). The NMM is a 5/7 multisig, who's members include the Noble corporate entity and core ecosystem partners. A key objective of the NMM is to achieve a reasonable level of geographic diversity for governance resilience. The current multisig members are listed below:

- Binary Holdings: Switzerland
- Chorus One: Switzerland
- Cosmostation: Korea
- Iqlusion: USA
- Luganodes: Switzerland
- Noble (the organization): Canada
- Strangelove: USA

The NMM is given the following key responsibilities:

**Upgrade Authority** 

The NMM has an on-chain procedure to instruct validators to coordinate a halt of the protocol running the Noble chain at a specific block height, to then restart the blockchain with a new version. Importantly, validators always have discretionary authority to follow those instructions. If 1/3 + 1 of the voting power on the Noble chain refuses to follow these upgrade instructions, the chain halts.

**Parameter Changes**

The NMM has the ability to initiate parameter changes. This authority is distinct from the upgrade functionality as it would not require the Noble validators to run a new chain binary. A parameter change is automatically be introduced into the state machine of the Noble chain.

**IBC Maintenance**

The NMM has the ability to both re-establish IBC connections that have expired and sever them in case of an emergency. Any such action related to an IBC connection is achieved automatically upon execution by the multisig.

## For Developers
 
### Installing Noble Locally

The following software is a prerequisite to installing Noble:

1. Go Programming Language (https://go.dev)
2. Git (https://git-scm.com)
3. GNU Make (https://www.gnu.org/software/make)

```
git clone https://github.com/noble-assets/noble
cd noble
make install
```

### Running a Local Network

To quickly spin up a standalone Noble chain locally and setup all privileged accounts, run the [local.sh](local.sh) script:

```
sh local.sh -r
```

## For Node Operators & Validators

See our [`networks`](https://github.com/noble-assets/networks) repository for all information around running Noble for mainnet and testnet!
