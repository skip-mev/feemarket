# Upgrading Your Chain to Use FeeMarket

## Code Changes

To integrate your chain with `x/feemarket`, the following steps should be performed:

### Add the Module to Your App

- The `FeeMarketKeeper` must be added to your application as seen [here](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/tests/app/app.go#L163).
- A `DenomResolver` (if desired) must be set in your application as seen [here](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/tests/app/app.go#L509).
- `Ante` and `Post` handlers must be configured and set with the application `FeeMarketKeeper` as seen [here](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/tests/app/app.go#L513).

### Determine Parameters

We recommend using 

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


