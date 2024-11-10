*Nov 10, 2024*

This is the Helium major release of Noble. It upgrades Noble itself and all of it's core modules to the latest stable release of the Cosmos SDK, `v0.50.x` a.k.a. Eden. Additional module changes have been documented below:

#### FiatTokenFactory

The BlockIBC logic was improved to support both Bech32 and Bech32m for IBC recipient addresses.

#### Florin

The module was updated to accept a user's public key when verifying signatures, instead of relying on on-chain data.

#### Forwarding

The module was updated to include a fallback address and a list of allowed denominations to forward.
