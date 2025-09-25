# Security

This document outlines Noble's security architecture at different layers of the blockchain stack.

## Generic Security Architecture

The first diagram shows Noble's overall security model, illustrating the separation between
consensus, core protocol, and application layers:

```mermaid
  flowchart LR
    subgraph Consensus
      bft[Validators]
    end

    subgraph Core
      n[Nova module]
    end

    subgraph AppLayer
      sc[Smart Contracts]
    end

    abci([ABCI 2.0])

    bft --> abci
    abci --> n

    bft --> AppLayer
```

## Token Issuance Security

The second diagram focuses specifically on token issuance security, showing how different authorities control different tokens while validators secure the overall network:

```mermaid
  flowchart LR
    subgraph Validators
     nmm[NMM]
    end

    subgraph Consensus
      Validators
    end

    subgraph TA[Token A]
     ipa[issuance properties]
    end

    subgraph TB[Token B]
     ipb[issuance properties]
    end

    subgraph Core
     authority
     TA
     TB
    end

    nmm -- maintenance --> Core
    Validators -- validate blocks --> Core

    authoritya[Authority A] -- controls --> TA
    authorityb[Authority B] -- controls --> TB
```
