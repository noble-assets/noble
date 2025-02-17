# Local Net Quick Start

Before running any of the scripts below, ensure you have built Noble using `make build`.

- [Single Validator Network](#single-validator-network)
- [Multi Validator Network](#multi-validator-network)
- [In-Place Testnet Fork](#in-place-testnet-fork)

## Single Validator Network

Start a single validator local Noble network.

```sh
sh single-val.sh -r
```

## Multi Validator Network

Start a three validator local Noble network.

Note: this requires [tmux](https://github.com/tmux/tmux/wiki).

```sh
sh multi-val.sh -r

# How to kill:
#   `ctr-c` kill 1 out of the three nodes
    killall nobled # kill remaining noble nodes
#   (`ctrl-b` then `d`) exit out of tmux session 
    tmux kill-session -t 3v-network # kill tmux session
```

## In-Place Testnet Fork

Synchronize a mainnet (or testnet) node using state sync, then create an `in-place-testnet`.

Note: your noble binary in the `build` folder must be compatible with the relevant network.

```sh
# ARGS:
#   -r|--reset                   - delete chain home folder resetting network
#   -t|--testnet                 - sync testnet instead of mainnet
#   -u|--trigger-testnet-upgrade - trigger an upgrade handler to run on the first block of the forked testnet

# mainnet example:
sh in-place-fork.sh -r

# testnet example:
sh in-place-fork.sh -r -t

# trigger upgrade example:
sh in-place-fork.sh -r -t -u "v9.0.0-rc.0"
```
