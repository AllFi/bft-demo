# An experimental application to demonstrate the capabilities of the Tendermint consensus protocol
This application has command line interface that allows to run Tendermint nodes with built-in simple ABCI apps and interact with them. 
The ABCI app expects integers as transactions. A correct node considers transactions with an even number to be valid, 
while a malicious node considers transactions with an odd number to be valid.

## Minimum requirements

| Requirement | Notes            |
|-------------|------------------|
| Go version  | Go1.15 or higher |

## Build
```
go install
```

## Example
Initialize four Tendermint validator nodes.
```
bft-demo init 4 
```

Run each node by index in separated console. The node is always correct after a startup.
```
bft-demo run 0 
bft-demo run 1 
bft-demo run 2 
bft-demo run 3 
```

Broadcast a transaction with the value 2 to the node with the index 0.
```
bft-demo tx 0 2
```

Make the node with the index 0 malicious.
```
bft-demo changeStatus 0 Malicious
```

Broadcast transaction with the value 3 to the node with the index 0. The node with the index 0 will panic as it recognizes the transaction as valid, while +2/3 nodes recognize the transaction as invalid.
```
bft-demo tx 0 3
```

The recover command can be used to restore the state of the malicious node by copying the state of one of the correct nodes. 
This requires stopping and restarting the node after recovery.
```
bft-demo recover 0 1
```

