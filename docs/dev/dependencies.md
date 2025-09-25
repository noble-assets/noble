# Dependencies

This document provides a visual representation of the Noble blockchain's modular architecture
and the dependencies between its various components.

## Module Dependency Graph

The diagram below shows the modules composing the Noble Core chain and their dependencies graph.
The modules are organized into three main categories:

- **Cosmos SDK**: Standard blockchain modules from the Cosmos SDK framework
- **Noble**: Custom modules specific to Noble's functionality
- **External**: Third-party modules integrated into Noble

Each arrow represents a dependency relationship, showing how modules depend on each other for functionality.

```mermaid
flowchart LR
   subgraph CM[Cosmos SDK]
      auth[auth]
      authz[authz]
      bank[bank]
      consensus[consensus]
      crisis[crisis]
      evidence[evidence]
      feegrant[feegrant]
      params[params]
      slashing[slashing]
      staking[staking]
      upgrade[upgrade]
  end

  subgraph Noble
      authority[authority]
      dollar[dollar]
      forwarding[forwarding]
      globalfee[globalfee]
      orbiter[orbiter]
      swap[swap]
      wormhole[wormhole]
  end

  subgraph External
      capability[capability]
      ibc[ibc]
      ica[interchainaccounts]
      pfm[packetforwardingmiddleware]
      ratelimit[ratelimit]
      transfer[ics20]

      cctp[cctp]
      ftf[fiat-tokenfactory]

      aura[aura]

      halo[halo]

      hyperlane[hyperlane]
      warp[warp]

      florin[florin]
  end

    authority --> auth
    authority --> bank

    authz --> bank
    authz --> auth

    dollar --> warp
    dollar --> wormhole
    dollar --> auth
    dollar --> bank

    forwarding --> bank
    forwarding --> transfer
    forwarding --> auth

    feegrant --> bank
    feegrant --> auth

    swap --> bank
    swap --> auth

    evidence --> staking
    evidence --> slashing

    ratelimit --> bank

    staking --> bank
    staking --> auth

    slashing --> staking

    halo --> bank
    halo --> auth

    aura --> bank

    florin --> bank

    cctp --> ftf
    cctp --> bank

    warp --> hyperlane
    warp --> bank

    hyperlane --> bank

    transfer --> ibc
    transfer --> bank
    transfer --> auth

    pfm --> bank
    pfm --> transfer

    orbiter --> bank
```
