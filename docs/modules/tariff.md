# Tariff module

The tariff module is meant to sit before the distribution module. This module collects a percentage of a specified asset and distributes it among entities.

There are several customizable parameters below:

## Parameters:

- `Share`: percentage of transaction to distribute among `DistributionEntities`

- `DistributionEntities`: Addresses that will acquire a specified percentage of the asset. The asset will be divided between each `DistributionEntity` based on their `Share` percentage. The `Share` across all `DistributionEntities` must equal `1`. Note that there are two "Shares"; the Tariff module Share and the each DistributionEntity Share.

- `TransferFeeBps`: Transfer Fee Basis Points is the model that calculates the total fee to distribute among `DistributionEntities`. TransferFeeBPS*10⁻⁴ = the total fee to distribute. 

- `TransferFeeMax`: The max amount that can be taken out of the transaction to distribute among `DistributionEntities`.

- `TransferFeeDenom`: The denom, if transacted, that will be distributed between `DistributionEntities`.


Example:
`Share`: 0.8

`DistributionEntities`: [("Jim", 0.3),("Mary", 0.7)]

`TransferFeeBps`: 1

`TransferFeeMax`: 5000000

`TransferFeeDenom`: stake


Alice sends 100stake to Bob. Since BPS is 1, the total fee collected is 0.01stake (100 * .0001). 

Since the `Share` percentage is 0.8, 80% of that 0.01stake will be divided among the entities. 80% of .0001 is 0.00008.

Jim will get 30% making his share 0.000024stake. Mary will get 70% making her share 0.000056stake.


--
Todo:
How/what will be handed off to the distribution model? 