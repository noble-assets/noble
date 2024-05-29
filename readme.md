![Noble banner](https://raw.githubusercontent.com/noble-assets/networks/main/Twitter_Banner.png)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fnoble-assets%2Fnoble.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fnoble-assets%2Fnoble?ref=badge_shield)
# ✨✨ Noble ✨✨

## Overview

[Noble](https://nobleassets.xyz/) is a Cosmos [application-specific blockchain](https://docs.cosmos.network/v0.46/intro/why-app-specific.html) purpose-built for native asset issuance. Noble brings the efficiency and interoperability of native assets to the wider Cosmos ecosystem, starting with USDC. Noble’s vision is to be the world’s premier issuance hub for digital assets that connect to other blockchains seamlessly. Noble leverages the Cosmos-SDK – a flexible toolkit that allows developers to leverage existing modules, and to seamlessly integrate custom modules that add virtually unlimited functionality for asset issuers onthe Noble blockchain.

## Noble App-Chain Design 

The Noble app-chain conforms to [industry standard](https://github.com/centrehq/centre-tokens/blob/master/doc/tokendesign.md) smart contracting capabilities with regards to asset issuance functionality. This functionality allows the minting and burning of tokens by multiple entities and the freezing blacklisting of addresses on the Noble chain.

Key authorities include: 

**Owner:** The Owner role has the ability to reassign all roles and will be held by the asset issuer. 

**Minter:** The Owner has the authority to add/remove Minters which have the authority to both mint and burn tokens on the Noble chain. 

**Blacklist:** The asset issuer has the authority to blacklist addresses. A blacklisted address will be unable to transfer tokens outside of the Noble chain via IBC, or to approve, mint, or burn tokens. 


## Tokenfactory

Noble implements a tokenfactory module pattern to enable the minting of generic assets in Cosmos. Further, tokenfactory modules are under the exclusive purview of the asset issuers and are separate and distinct from governance of the Noble chain. Additionally, each tokenfactory module is distinct from one other as they each house unique access controls with ownership over the minting and burning of a specific asset.

The Access Control table below shows the functionality tied to each privileged account.

## Access Control

|                                             | **Owner** | **Minter** | **Master Minter** | **Minter Controller** | **Pauser** | **Blacklister** | **Is Paused<br>(Actions Allowed)** |
|---------------------------------------------|:---------:|:----------:|:-----------------:|:---------------------:|:----------:|:---------------:|:--------------------------------:|
| **Blacklist**                               |           |            |                   |                       |            |        x        |                 x                |
| **Unblacklist**                             |           |            |                   |                       |            |        x        |                 x                |
| **Burn**                                    |           |      x     |                   |                       |            |                 |                                  |
| **Mint**                                    |           |      x     |                   |                       |            |                 |                                  |
| **Add Minter Controller**                   |           |            |         x         |                       |            |                 |                 x                |
| **Add Minter with allowance**               |           |            |                   |           x           |            |                 |                 x                |
| **Pause**                                   |           |            |                   |                       |      x     |                 |                 x                |
| **Unpause**                                 |           |            |                   |                       |      x     |                 |                 x                |
| **Remove Minter Controller**                |           |            |         x         |                       |            |                 |                 x                |
| **Remove Minter**                           |           |            |                   |           x           |            |                 |                 x                |
| **Update Blacklister**                      |     x     |            |                   |                       |            |                 |                 x                |
| **Update Master Minter**                    |     x     |            |                   |                       |            |                 |                 x                |
| **Update Owner**                            |     x     |            |                   |                       |            |                 |                 x                |
| **Update Pauser**                           |     x     |            |                   |                       |            |                 |                 x                |
| **Transer Tokens (tokenfactory asset)**     |     x     |      x     |         x         |           x           |      x     |        x        |                                  |
| **Transer Tokens (non-tokenfactory asset)** |     x     |      x     |         x         |           x           |      x     |        x        |                 x                |


## Security Guarantees in a Permissioned Validator Set (Proof of Authority) 

The initial plans for Noble’s go-to-market entailed leveraging Replicated Security (also known as ["Interchain Security”](https://github.com/cosmos/interchain-security)) – the shared security model of the Cosmos Hub. This feature allows validators to secure app-chains using the value of ATOMs staked by Cosmos Hub delegators. Due to uncertainty around the design of replicated security (e.g., the removal of automated slashing packets), Noble has paused integration with Replicated Security. 

At launch, Noble will be a Proof of Authority chain with a trusted validator set (a subset of Cosmos Hub validators). The validator set will be permissioned by equal shares of staking tokens placed in vesting accounts. The tokens will have no value. Economic security will derive from fees captured by the chain in USDC and other assets on a block-by-block basis. If a double sign is detected by the chain, the validator address will be “tombstoned,” meaning that their tokens and the address will no longer be usable for validation and resulting in the loss of all future fee revenue. The Proof of Authority model provides real economic cost to faulty validator behavior and thus provides economic security to the network that can be computed in real time based on past and projected fees.

Noble intends to monitor developments in shared security across the blockchain ecosystem to ensure the optimality of the security model.

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
- Strangelove: USA 
- Chorus One: Switzerland
- Binary Holdings: Switzerland 
- Cosmostation: Korea  
- Luganodes: Switzerland 

Stakeholders on the NMM will be able to be rotated in and out of the mulitisg over time. 

## Partnership with StrangeLove & Future Technical Development

Noble has selected StrangeLove Labs (“StrangeLove”) as its initial engineering and product partner to implement Noble’s initial app-chain design. StrangeLove has built an MVP and is bringing it to market using battle-proven XP-derived practices like Test-driven Development, Continuous Integration and Continuous Deployment, and short iteration cycles.

In collaboration with StrangeLove, Noble intends to build an in-house team focused on software development and integration to further build the Noble vision related to products and services.

## Media

TWITTER: [@noble_xyz](https://twitter.com/noble_xyz/)

WEB: [nobleassets.xyz](https://nobleassets.xyz/)

 
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

Prior to launch, some config settings are required by the `tokenfactory` module in the `genesis.json` file:
- There must be a designated address for the "owner" role.
- Other roles, such as `masterMinter`, `blacklister`, and `pauser` are optional at genesis.
- `mintingDenom` must be filled out and the `denom` specified must be registered in `denom_metadata`

`tokenfactory `Example:
```json
"tokenfactory": {
      "params": {},
      "blacklistedList": [],
      "paused": {
        "paused": false
      },
      "masterMinter": {
        "address": "noble1x8rynykqla7cnc0tf2f3xn0wa822ztt70y39vn"
      },
      "mintersList": [],
      "pauser": {
        "address": "noble1g3v4qdc83h6m5wdz3x92vfu0tjtt7e6y48qqrz"
      },
      "blacklister": {
        "address": "noble159leclhhuhhcmedu2n8nfjjedxjyrtkee8l4v2"
      },
      "owner": {
        "address": "noble153eyy4uufmrak2swgrn4fjtyslg256ecdngyve"
      },
      "minterControllerList": [],
      "mintingDenom": {
        "denom": "urupee"
      }
    }
```

`denom_metadata` example:

```json
"denom_metadata": [
        {
          "display": "rupee",
          "base": "urupee",
          "name": "rupee",
          "symbol": "rupee",
          "denom_units": [
            {
              "denom": "urupee",
              "aliases": [
                "microrupee"
              ],
              "exponent": "0"
            },
            {
              "denom": "mrupee",
              "aliases": [
                "milirupee"
              ],
              "exponent": "3"
            },
            {
              "denom": "rupee",
              "aliases": null,
              "exponent": "6"
            }
          ]
        },
    ]
```
## Launch node

```
nobled start
```

## Running Nodes

- See [`noble-networks`](https://github.com/noble-assets/networks/tree/main) for all info around validating for mainnet and testent

## Testing

The [`interchaintest`](https://github.com/strangelove-ventures/interchaintest) test suite has been imported into the noble repo. Tests can be ran and written [here](./interchaintest/).

To quickly spin up a standalone noble chain and setup all privileged accounts, run the [play.sh](play.sh) bash script.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fnoble-assets%2Fnoble.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fnoble-assets%2Fnoble?ref=badge_large)