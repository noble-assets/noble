# Tariff module

The tariff module is meant to sit before the distribution module in the begin-block sequence. This module collects a percentage of a specified asset and distributes it among configured entities.


## Parameters:

- `Share`: percentage of collected fees to distribute among `DistributionEntities`

- `DistributionEntities`: Addresses that will acquire a specified percentage of the overall `Share`. The collected fees will be divided between each `DistributionEntity` based on their individual `Share` percentage. The sum of the `Share` across all `DistributionEntities` must equal `1`. Note that there are two `Share`s; the Tariff module overall `Share` and the `Share` for each `DistributionEntity`.

- `TransferFeeBps`: Transfer Fee Basis Points (BPS) is the parameter that determines the BPS fees to be collected for outgoing IBC transfers, up to the `TransferFeeMax`, for the `TransferFeeDenom`. This fee is collected in addition to the transaction gas fees. `TransferFeeBPS`*10⁻⁴ = the fee multiplier applied to the outgoing transfer amount.

- `TransferFeeMax`: The max amount of fees to be collected for an outgoing IBC transfer.

- `TransferFeeDenom`: The denom to collect fees for on outgoing IBC transfers.

---

## Example

`Share`: 0.8

`DistributionEntities`: "Jim" has a  30% share, "Mary" has a 70%

`TransferFeeBps`: 1

`TransferFeeMax`: 5000000

`TransferFeeDenom`: ustake

For sake of example, lets assume gas prices are 0.

Alice sends 100_000_000ustake to Bob on a different chain using IBC. Since BPS is 1, the total fee collected is 10_000ustake (100_000_000 * .0001). 

Since the `Share` percentage is 0.8, 80% of that 10_000ustake will be divided among the entities. 80% of 10_000 is 8_000.

Jim will get 30% of the 8_000 making his share 2_400ustake. Mary will get 70% making her share 5_600ustake.

The remaining 2_000ustake of collected fees, which were not distributed among the `DistributionEntities`, are distributed by the distribution module. The distribution module will distribute the 2_000ustake among the validators weighted by their voting power. For Noble's Proof of Authority (POA) use case, all validators have equal voting power. As such, the 2_000ustake is divided equally among the validators. 

Since the distribution logic truncates to the nearest integer, fees can be left over after the distribution module. This is expected behavior as fees will be distributed again in the next block.

When gas prices are non-zero, the fees collected are distributed in the same way: tariff module distribution entities first, then distribution module to the validators. 