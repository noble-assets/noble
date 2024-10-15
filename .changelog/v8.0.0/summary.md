*October 17, 2024* 

This is the Helium major release of Noble. It upgrades the Noble
binary to Cosmos SDK version 50 and IBC-Go to version 8.
In addition, it replaces the Paramauthority module with an in-house built
[Authority][authority] module. This module is used to assist with governance 
tasks such as chain upgrades and IBC client substitutions. 

The following Noble specific modules have been updated to SDK v50 and 
IBC-Go v8:

- [cctp]
- [fiat-tokenfactory]
- [aura]
- [halo]
- [florin]
- [forwarding]

[authority]: https://github.com/noble-assets/authority
[cctp]: https://github.com/circlefin/noble-cctp
[fiat-tokenfactory]: https://github.com/circlefin/noble-fiattokenfactory
[aura]: https://github.com/ondoprotocol/usdy-noble
[halo]: https://github.com/noble-assets/halo
[florin]: https://github.com/monerium/module-noble
[forwarding]: https://github.com/noble-assets/forwarding
