# `x/feemarket`

## Abstract

This document specifies the feemarket module.

The feemarket module is an implementation of the Additive Increase Multiplicative Decrease (AIMD) EIP-1559 
feemarket.  More information about the implementation can be found [here](./x/feemarket/README.md).

This module is planned to be used in the Cosmos Hub.

## Contents

* [State](#state)
  * [BaseFee](#basefee)
  * [LearningRate](#learningrate)
  * [Window](#window)
  * [Index](#index)
* [Keeper](#keeper)
* [Messages](#messages)
* [Events](#events)
    * [FeePay](#feepay)
    * [TipPay](#tippay)
* [Parameters](#parameters)
    * [Alpha](#alpha)
    * [Beta](#beta)
    * [Theta](#theta)
    * [Delta](#delta)
    * [MinBaseFee](#minbasefee)
    * [MinLearningRate](#minlearningrate)
    * [MaxLearningRate](#maxlearningrate)
    * [TargetBlockUtilization](#targetblockutilization)
    * [MaxBlockUtilization](#maxblockutilization)
    * [Window](#window)
    * [FeeDenom](#feedenom)
    * [Enabled](#enabled)
* [Client](#client)
    * [CLI](#cli)
    * [Query](#query)
* [gRPC](#grpc)

## State

The `x/feemarket` module keeps state of the following primary objects:

1. Current base-fee
2. Current learning rate
3. Moving window of block utilization

In addition, the `x/feemarket` module keeps the following indexes to manage the
aforementioned state:

* State: `0x02 |ProtocolBuffer(State)`

### BaseFee

BaseFee is the current base fee. This is denominated in the fee per gas
unit.

### LearningRate

LearningRate is the current learning rate.

### Window

Window contains a list of the last blocks' utilization values. This is used
to calculate the next base fee. This stores the number of units of gas
consumed per block.

### Index

Index is the index of the current block in the block utilization window.

```protobuf
// State is utilized to track the current state of the fee market. This includes
// the current base fee, learning rate, and block utilization within the
// specified AIMD window.
message State {
  // BaseFee is the current base fee. This is denominated in the fee per gas
  // unit.
  string base_fee = 1 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "cosmossdk.io/math.Int",
    (gogoproto.nullable) = false
  ];

  // LearningRate is the current learning rate.
  string learning_rate = 2 [
    (cosmos_proto.scalar) = "cosmos.Legacy",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Window contains a list of the last blocks' utilization values. This is used
  // to calculate the next base fee. This stores the number of units of gas
  // consumed per block.
  repeated uint64 window = 3;

  // Index is the index of the current block in the block utilization window.
  uint64 index = 4;
}
```

## Keeper

The feemarket module provides a keeper interface for accessing the KVStore.

```go
type FeeMarketKeeper interface {
	// Get the current state from the store.
    GetState(ctx sdk.Context) (types.State, error)

    // Set the state in the store.
    SetState(ctx sdk.Context, state types.State) error

    // Get the current params from the store.
    GetParams(ctx sdk.Context) (types.Params, error)

    // Set the params in the store.
    SetParams(ctx sdk.Context, params types.Params) error
	
	// Get the current minimum gas prices (base fee) from the store.
    GetMinGasPrices(ctx sdk.Context) (sdk.Coins, error)
}
```

## Messages

### MsgParams

The `feemarket` module params can be updated through `MsgParams`, which can be done using a governance proposal. The signer will always be the `gov` module account address.

```protobuf
message MsgParams {
  option (cosmos.msg.v1.signer) = "authority";

  // Params defines the new parameters for the feemarket module.
  Params params = 1 [ (gogoproto.nullable) = false ];
  // Authority defines the authority that is updating the feemarket module
  // parameters.
  string authority = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
}
```

The message handling can fail if:

* signer is not the gov module account address.


## Events

The feemarket module emits the following events:

#### FeePay

```json
{
  "type": "fee_pay",
  "attributes": [
    {
      "key": "fee",
      "value": "{{sdk.Coins being payed}}",
      "index": true
    },
    {
      "key": "fee_payer",
      "value": "{{sdk.AccAddress paying the fees}}",
      "index": true
    }
  ]
}
```

#### TipPay

```json
{
  "type": "tip_pay",
  "attributes": [
    {
      "key": "tip",
      "value": "{{sdk.Coins being payed}}",
      "index": true
    },
    {
      "key": "tip_payer",
      "value": "{{sdk.AccAddress paying the tip}}",
      "index": true
    },
    {
      "key": "tip_payee",
      "value": "{{sdk.AccAddress receiving the tip}}",
      "index": true
    }
  ]
}
```

## Parameters

The feemarket module stores it's params in state with the prefix of `0x01`,
which can be updated with governance or the address with authority.

* Params: `0x01 | ProtocolBuffer(Params)`

The feemarket module contains the following parameters:

### Alpha

Alpha is the amount we added to the learning rate
when it is above or below the target +/- threshold.

### Beta

The default send enabled value controls send transfer capability for all
coin denominations unless specifically included in the array of `SendEnabled`
parameters.

### Theta

Theta is the threshold for the learning rate. If the learning rate is
above or below the target +/- threshold, we additively increase the
learning rate by Alpha. Otherwise, we multiplicatively decrease the
learning rate by Beta.

### Delta

Delta is the amount we additively increase/decrease the base fee when the
net block utilization difference in the window is above/below the target
utilization.

### MinBaseFee

MinBaseFee determines the initial base fee of the module and the global
minimum for the network. This is denominated in fee per gas unit.

### MinLearningRate

MinLearningRate is the lower bound for the learning rate.

### MaxLearningRate

MaxLearningRate is the upper bound for the learning rate.

### TargetBlockUtilization

TargetBlockUtilization is the target block utilization for the current block.

### MaxBlockUtilization

MaxBlockUtilization is the maximum block utilization.  Once this has been surpassed,
no more transactions will be added to the current block.

### Window

Window defines the window size for calculating an adaptive learning rate
over a moving window of blocks.  The default EIP1559 implementation uses
a window of size 1.

### FeeDenom

FeeDenom is the denom that will be used for all fee payments.

### Enabled

Enabled is a boolean that determines whether the EIP1559 fee market is
enabled. This can be used to add the feemarket module and enable it 
through governance at a later time.

```protobuf
// Params contains the required set of parameters for the EIP1559 fee market
// plugin implementation.
message Params {
  // Alpha is the amount we additively increase the learninig rate
  // when it is above or below the target +/- threshold.
  string alpha = 1 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Beta is the amount we multiplicatively decrease the learning rate
  // when it is within the target +/- threshold.
  string beta = 2 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Theta is the threshold for the learning rate. If the learning rate is
  // above or below the target +/- threshold, we additively increase the
  // learning rate by Alpha. Otherwise, we multiplicatively decrease the
  // learning rate by Beta.
  string theta = 3 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // Delta is the amount we additively increase/decrease the base fee when the
  // net block utilization difference in the window is above/below the target
  // utilization.
  string delta = 4 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // MinBaseFee determines the initial base fee of the module and the global
  // minimum
  // for the network. This is denominated in fee per gas unit.
  string min_base_fee = 5 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "cosmossdk.io/math.Int",
    (gogoproto.nullable) = false
  ];

  // MinLearningRate is the lower bound for the learning rate.
  string min_learning_rate = 6 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // MaxLearningRate is the upper bound for the learning rate.
  string max_learning_rate = 7 [
    (cosmos_proto.scalar) = "cosmos.Dec",
    (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
    (gogoproto.nullable) = false
  ];

  // TargetBlockUtilization is the target block utilization.
  uint64 target_block_utilization = 8;

  // MaxBlockUtilization is the maximum block utilization.
  uint64 max_block_utilization = 9;

  // Window defines the window size for calculating an adaptive learning rate
  // over a moving window of blocks.
  uint64 window = 10;

  // FeeDenom is the denom that will be used for all fee payments.
  string fee_denom = 11;

  // Enabled is a boolean that determines whether the EIP1559 fee market is
  // enabled.
  bool enabled = 12;
}
```

## Client

### CLI

A user can query and interact with the `feemarket` module using the CLI.

#### Query

The `query` commands allow users to query `feemarket` state.

```shell
feemarketd query feemarket --help
```

##### params

The `params` command allows users to query the on-chain parameters.

```shell
feemarketd query feemarket params [flags]
```

Example:

```shell
feemarketd query feemarket params
```

Example Output:

```yml
alpha: "0.000000000000000000"
beta: "1.000000000000000000"
delta: "0.000000000000000000"
enabled: true
fee_denom: stake
max_block_utilization: "30000000"
max_learning_rate: "0.125000000000000000"
min_base_fee: "1000000"
min_learning_rate: "0.125000000000000000"
target_block_utilization: "15000000"
theta: "0.000000000000000000"
window: "1"
```

##### state

The `state` command allows users to query the current on-chain state.

```shell
feemarketd query feemarket state [flags]
```

Example:

```shell
feemarketd query feemarket state
```

Example Output:

```yml
base_fee: "1000000"
index: "0"
learning_rate: "0.125000000000000000"
window:
  - "0"
```

##### base-fee

The `base-fee` command allows users to query the current base-fee.

```shell
feemarketd query feemarket base-fee [flags]
```

Example:

```shell
feemarketd query feemarket base-fee
```

Example Output:

```yml
1000000stake
```

## gRPC

A user can query the `feemarket` module using gRPC endpoints.

### Params

The `Params` endpoint allows users to query the on-chain parameters.

```shell
feemarket.feemarket.v1.Query/Params
```

Example:

```shell
grpcurl -plaintext \
    localhost:9090 \
    feemarket.feemarket.v1.Query/Params
```

Example Output:

```json
{
  "params": {
    "alpha": "0",
    "beta": "1000000000000000000",
    "theta": "0",
    "delta": "0",
    "minBaseFee": "1000000",
    "minLearningRate": "125000000000000000",
    "maxLearningRate": "125000000000000000",
    "targetBlockUtilization": "15000000",
    "maxBlockUtilization": "30000000",
    "window": "1",
    "feeDenom": "stake",
    "enabled": true
  }
}
```

### State

The `State` endpoint allows users to query the current on-chain state.

```shell
feemarket.feemarket.v1.Query/State
```

Example:

```shell
grpcurl -plaintext \
    localhost:9090 \
    feemarket.feemarket.v1.Query/State
```

Example Output:

```json
{
  "state": {
    "baseFee": "1000000",
    "learningRate": "125000000000000000",
    "window": [
      "0"
    ]
  }
}
```

### BaseFee

The `BaseFee` endpoint allows users to query the current on-chain base-fee.

```shell
feemarket.feemarket.v1.Query/BaseFee
```

Example:

```shell
grpcurl -plaintext \
    localhost:9090 \
    feemarket.feemarket.v1.Query/BaseFee
```

Example Output:

```json
{
  "fees": [
    {
      "denom": "stake",
      "amount": "1000000"
    }
  ]
}
```

