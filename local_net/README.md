# Local Net Quick Start

Before running any of the scripts below, ensure you have built Noble using `make build`.

- [1 Val Network](#1-val-network)
- [3 Val Network](#3-val-network)
- [In-place testnet](#in-place-testnet)

## 1 Val Network

Start a single validator local Noble network.

```sh
sh ./local.sh -r
```

## 3 Val Network

Start a three validator local Noble network.

Note: this requries [tmux](https://github.com/tmux/tmux/wiki).

```sh
sh ./local_3val.sh -r

# How to kill:
#   `ctr-c` kill 1 out of the three nodes
    killall nobled # kill remaining noble nodes
#   (`ctrl-b` then `d`) exit out of tmux session 
    tmux kill-session -t 3v-network # kill tmux session
```

## In-Place Testnet

Synchronize a mainnet (or testnet) node using state sync, then create an `in-place-testnet`.

Note: your noble binary in the `build` folder must be compatible with the relevant network.

```sh
#mainnet
sh local_in_place.sh -r

#testnet
sh local_in_place.sh -r -t
```
