# Upgrading Your Chain to Use FeeMarket

## Code Changes

To integrate your chain with `x/feemarket`, the following steps should be performed:

### Add the Module to Your App

### Determine Parameters

### Upgrade Handlers

## Changes for End-Users

With the addition of `x/feemarket`, there are some important changes that end-users must be aware of.

1. A non-zero fee is _always required_ for all transactions.
   1. Pre-`x/feemarket` validators were able to set their `MinGasPrice` field locally, meaning it was possibly for some to have no required fees.  This is no longer true as there is always a non-zero global fee for transactions.
2. Fees are no longer static.
   1. The `gas price` will change with market activity, so to ensure that transactions will be included, wallets, relayers, etc. will need to query `x/feemarket` for the current fee state.  See the [Querying Gas Price](#querying-gas-price-) section below.
3. Fees _must_ always be a single coin denomination.
   1. Example `--fees skip` is valid while `--fees 10stake,10skip` is invalid 

Note that fees are still paid using the `fees` [field](https://github.com/cosmos/cosmos-sdk/blob/d1aab15790570bff77aa0b8652288a276205efb0/proto/cosmos/tx/v1beta1/tx.proto#L214) of a Cosmos SDK Transaction as they were before.

### Querying Gas Price 


