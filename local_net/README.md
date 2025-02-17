# Local Net Quick Start

Before running any of the scripts below, ensure you have built Noble using `make build`.

- [Single Validator Network](#sing-validator-network)
- [3 Val Network](#multi-validator-network)
- [In-place Testnet Fork](#in-place-testnet-fork)

## Sing Validator Network

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
#mainnet
sh in-place-fork.sh -r

#testnet
sh in-place-fork.sh -r -t
```
