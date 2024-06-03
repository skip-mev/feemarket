# Upgrading Your Chain to Use FeeMarket

## Code Changes

To integrate your chain with `x/feemarket`, the following steps should be performed:

### Add the Module to Your App

* The `FeeMarketKeeper` must be added to your application as seen [here](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/tests/app/app.go#L163).
* A `DenomResolver` (if desired) must be set in your application as seen [here](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/tests/app/app.go#L509).
* `Ante` and `Post` handlers must be configured and set with the application `FeeMarketKeeper` as seen [here](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/tests/app/app.go#L513).

### Determine Parameters

We provide sensible default parameters for running either the [EIP-1559](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/x/feemarket/types/eip1559.go#L56) or [AIMD EIP-1559](https://github.com/skip-mev/feemarket/blob/0f83e172c92a02db45f83bf89065fd9543967729/x/feemarket/types/eip1559_aimd.go#L65) feemarkets. 

> **Note**
>
> The default parameters use the default Cosmos SDK bond denomination. The should be modified to your chain's fee denomination.

## Changes for End-Users

With the addition of `x/feemarket`, there are some important changes that end-users must be aware of.

1. A non-zero fee is _always required_ for all transactions.
   1. Pre-`x/feemarket` validators were able to set their `MinGasPrice` field locally, meaning it was possibly for some to have no required fees.  This is no longer true as there is always a non-zero global fee for transactions.
2. Fees are no longer static.
   1. The `gas price` will change with market activity, so to ensure that transactions will be included, wallets, relayers, etc. will need to query `x/feemarket` for the current fee state.  See the [Querying Gas Price](#querying-gas-price-) section below.
3. Fees _must_ always be a single coin denomination.
   1. Example `--fees skip` is valid while `--fees 10stake,10skip` is invalid 

> **Note**
>
>  Fees are still paid using the `fees` [field](https://github.com/cosmos/cosmos-sdk/blob/d1aab15790570bff77aa0b8652288a276205efb0/proto/cosmos/tx/v1beta1/tx.proto#L214) of a Cosmos SDK Transaction as they were before.

### Querying Gas Price 

Extensive querying information can be seen in the module [spec](./README.md#query).

The specific query for `GasPrices` can be found [here](./README.md#gas-prices).

#### Code Snippet

Wallet, relayers, and other users will want to add programmatic ways to query this before building their transactions.  Below is an example of how a user could implement this lightweight query in Go:

```go
package example

import (
    "context"
	
    sdk "github.com/cosmos/cosmos-sdk/types"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

const endpoint = "TEST-ENDPOINT"

func QueryGasPrice(ctx context.Context, denom string) (sdk.DecCoin, error) {
	// Set up gRPC connection to chain
   cc, err := grpc.NewClient(endpoint, insecure.NewCredentials())
   if err != nil {
	   return sdk.DecCoin{}, err
   }
   defer cc.Close()
   
   // Create FeeMarketClient with underlying gRPC connection
   feeMarketClient := feemarkettypes.NewQueryClient(cc)
   
   gasPrice, err := feeMarketClient.GasPrice(ctx, &feemarkettypes.GasPriceRequest{
	   Denom: denom,
   })
   if err != nil {
	   return sdk.DecCoin{}, err
   }
   
   return gasPrice, nil
}
```

#### Examples of Other EIP-1559 Integrations

The [Osmosis](https://github.com/osmosis-labs/osmosis) Blockchain has a similar EIP-1559 feemarket that has been integrated by wallets and relayers.  Below are some examples as to how different projects query the dynamic fee for transactions:

* [Keplr Wallet EIP-1559 BaseFee Query](https://github.com/chainapsis/keplr-wallet/blob/b0a96c2c713d8163ce840fcd5abbac4eb612607c/packages/stores/src/query/osmosis/base-fee/index.ts#L18)
* [Cosmos-Relayer EIP-1559 BaseFee Query](https://github.com/cosmos/relayer/blob/9b140b664fe6b10161af1093ccd26627b942742e/relayer/chains/cosmos/fee_market.go#L13)
* [Hermes Relayer EIP-1559 Fee Query](https://github.com/informalsystems/hermes/blob/fc8376ba98e4b595e446b366b736a0c046d6026a/crates/relayer/src/chain/cosmos/eip_base_fee.rs#L15)
  * Note: Hermes also already implements a query `x/feemarket` seen [here](https://github.com/informalsystems/hermes/blob/fc8376ba98e4b595e446b366b736a0c046d6026a/crates/relayer/src/chain/cosmos/eip_base_fee.rs#L33)
