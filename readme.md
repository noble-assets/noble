# ✨✨ Noble ✨✨

## Overview

Noble is a Cosmos [application-specific blockchain](https://docs.cosmos.network/v0.46/intro/why-app-specific.html) purpose-built for native asset issuance for the wider Cosmos and IBC ecosystem. Noble leverages the [Cosmos-SDK](https://v1.cosmos.network/sdk) – a flexible toolkit that allows developers to leverage existing [modules](https://github.com/cosmos/cosmos-sdk), and to seamlessly integrate custom modules that add to the offerings of the Noble blockchain. 

## Noble App-Chain Design 

The Noble app-chain conforms to [industry standard](https://github.com/centrehq/centre-tokens/blob/master/doc/tokendesign.md) smart contracting capabilities with regards to asset issuance functionality. This functionality allows the minting and burning of tokens by multiple entities and the freezing of addresses on the Noble chain. 

Key authorities include: 

**Owner:** The Owner role has the ability to reassign all roles and will be held by the asset issuer. 

**Minter:** The Owner has the authority to add/remove Minters which have the authority to both mint and burn tokens on the Noble chain. 

**Blacklist:** The asset issuer has the authority to blacklist addresses. A blacklisted address will be unable to transfer tokens outside of the Noble chain via IBC, or to approve, mint, or burn tokens. 

Noble leverages [TokenFactory](https://docs.osmosis.zone/osmosis-core/modules/tokenfactory/) to enable the minting of generic assets in Cosmos. Further, TokenFactory modules are in the purview of the asset issuers and are distinct from governance of the app-chain. Additionally, each TokenFactory module is distinct from one other as they each house unique access controls with ownership over the minting and burning of a specific asset. 

## Tokenfactory

The Tokenfactory Module allows generic assets to be minted and controlled by privileged accounts.

The Access Control table below shows the functionality tied to each privileged account.

## Access Control

|                                | **Owner** | **Minter** | **Master Minter** | **Minter Controller** | **Pauser** | **Blacklister** | **Is Paused<br>(Actions Allowed)** |
|--------------------------------|:---------:|:----------:|:-----------------:|:---------------------:|:----------:|:---------------:|:--------------------------------:|
| **Blacklist**                  |           |            |                   |                       |            |        x        |                 x                |
| **Unblacklist**                |           |            |                   |                       |            |        x        |                 x                |
| **Burn**                       |           |      x     |                   |                       |            |                 |                                  |
| **Mint**                       |           |      x     |                   |                       |            |                 |                                  |
| **Configure Mint Controller**  |           |            |         x         |                       |            |                 |                 x                |
| **Configure Minter allowance** |           |            |                   |           x           |            |                 |                 x                |
| **Pause**                      |           |            |                   |                       |      x     |                 |                 x                |
| **Unpause**                    |           |            |                   |                       |      x     |                 |                 x                |
| **Remove Minter Controller**   |           |            |         x         |                       |            |                 |                 x                |
| **Remove Minter**              |           |            |                   |           x           |            |                 |                 x                |
| **Update Blacklister**         |     x     |            |                   |                       |            |                 |                 x                |
| **Update Master Minter**       |     x     |            |                   |                       |            |                 |                 x                |
| **Update Owner**               |     x     |            |                   |                       |            |                 |                 x                |
| **Update Pauser**              |     x     |            |                   |                       |            |                 |                 x                |
| **Transer Tokens**             |     x     |      x     |         x         |           x           |      x     |        x        |                                  |

## Security Guarantees in a Permissioned Validator Set (Proof of Authority) 

Initial plans for Noble’s go-to-market entailed leveraging Replicated Security (also known as ["Interchain Security”](https://github.com/cosmos/interchain-security)) – the shared security model of the Cosmos Hub. This feature allows validators to secure app-chains using the Hub’s billions of dollars of staked ATOMs. *As a result of the [latest updates](https://forum.cosmos.network/t/slashing-updates-in-replicated-security/9571) to the Interchain Security design in February 2023, notably the removal of automated slashing packets, Noble has decided to pause integrations of the Cross Chain Validation (CCV) protocol.* 

As such, Noble will launch a Proof of Authority chain with a trusted validator set (a subset of the Cosmos Hub). The validator set will be permissioned by equal shares of staking tokens placed in vesting accounts. The tokens will have no value. Even though the tokens have no value, there will be economic security. The validators will receive fees captured by the chain in USDC and other assets on a block by block basis. If a double sign is detected by the chain, the validors address will be “tombstoned”. This means that their tokens and the address will no longer be usable for validation. This will result in the loss of all future fee revenue.

This provides real economic cost to faulty validator behavior and thus provides economic security to the network that can be computed in real time based on past and projected fees.

## Connecting to Noble

App-chains are able to permissionlessly connect to Noble [via IBC](https://medium.com/the-interchain-foundation/eli5-what-is-ibc-def44d7b5b4c), a universal interoperability protocol that allows two different blockchains to communicate with one another, garaunteeing reliable, ordered, and authenticated communication.

How to integrate with Noble? 

To establish a connection to Noble (or any 2 IBC compatible chains), developers will be required to set up a relayer on an existing path. This [tutorial](https://github.com/cosmos/relayer/blob/main/docs/create-path-across-chain.md) gives an overview of how to create a channel, client, and connection and to start relaying IBC packets. 

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
- StrangelLove: USA 
- Figment: Canada
- Binary Holdings: Switzerland 
- Cosmostation: Korea  
- EverStake: Ukraine/ Austria 

Stakeholders on the NMM will be able to be rotated in and out of the mulitisg over time. 

## Partnership with StrangeLove & Future Technical Development

Noble is working closely with the StrangeLove product and engineering team to implement the initial app-chain design. Strangelove works to build an MVP and bring it to market using battle-proven XP-derived practices like Test-driven Development, Continuous Integration and Continuous Deployment, and short iteration cycles.

As the collaboration brings the software to market, Strangelove will also work with the Noble team to interview and hire an engineering team independent of StrangeLove. Over time, Strangelove passes off development and operational control, so that when the Strangelove partnership eventually tapers off, the Noble team will house the majority of software development and technical expertise of Noble’s product and services offering. 

## Media

TWITTER: @noble_xy 

COMING SOON: nobleassets.xyz 

 
## Build and install to go bin path

[go](https://go.dev/dl/) will be needed for this install

```
make install
```

## Initialize config

Come up with a moniker for your node, then run:

```
nobled init $MONIKER
```

## Launch with genesis file or run as standalone chain

To launch as a consumer chain, download and save shared genesis file to [TODO: Path to Genesis]. Additionally add peering information (`persistent_peers` or `seeds`) to `~/.noble/config/config.toml`

## Launch node

```
nobled start
```

## Testing

Testing Noble as a standalone chain (without Provider) is possible by running the below command prior to starting node:

```
nobled add-consumer-section
```

To quickly spin up a standalone noble chain and setup all privileged accounts, refer to[ play.sh](play.sh)
