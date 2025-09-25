# Issuance

This document shows the token issuance ecosystem on Noble, illustrating the relationship
between issuers, tokens, and the Noble modules that manage them.

## Token Issuance Flow

The diagram below shows how various issuers deploy different tokens on Noble and which
specialized modules manage each token type:

- **Issuers**: Real-world entities that mint and manage stablecoins.

- **Tokens**: The actual digital assets issued (USDC, USDN, EURe, etc.).

- **Modules**: Noble's specialized modules that handle the on-chain logic for each token type.

All token modules ultimately interact with the Cosmos SDK `bank` module for core token
functionality like transfers and balances.

```mermaid
 flowchart TD
    subgraph Issuers
    m0[M^0]
    circle[Circle]
    monereum[Monereum]
    hash[Hashnode]
    ondo[Ondo]
    end

    subgraph tokens
      usdn((USDN))
      usdc((USDC))
      eure((EURe))
      usdy((USDY))
      usyc((USYC))
    end

    subgraph modules
      dollar[dollar]
      ftf[fiat-tokenfactory]
      aura[aura]
      halo[halo]
      florin[florin]
    end

    m0 --> usdn
    circle --> usdc
    monereum --> eure
    hash --> usyc
    ondo --> usdy


    usdn --> dollar
    usdc --> ftf
    usdy --> aura
    eure --> florin
    usyc --> halo

    dollar --> bank
    ftf --> bank
    aura --> bank
    halo --> bank
    florin --> bank
```
