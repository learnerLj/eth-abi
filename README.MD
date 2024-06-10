

# generate graph
## run
The entrance is at the main function in `main.go` where it reads function signature, event signature and ABI database and uses them to decode function parameters, logs and function outputs.
`fromDir` is the folder of your labeled transactions. `LoadFunctionSignatures`, `LoadEventSignatures`, `LoadABIs` are the mentioned databases. Then, `toDir` shows the location to store your result where the file name would remain same as it in `dir`.
```go
	fromDir := "./datum/txs/labeled"
	toDir := "./datum/txs/results"

	sigDb := "./datum/dbs/func_sig.json"
	eventDb := "./datum/dbs/events_sig.json"
	abiDb := "./datum/dbs/addr2abi.json"

```

you should first run `go mod tidy` at the root of this repo to makr sure all dependencies ready. Then adjust the beforementioned parameters to fit the path of your databases. Afterward, `go run ./main.go`. it may take a while to load databases.

## format of inputs and outputs
In the input files, key is the transaction hash and the value is standard results of ethereum transaction recipient, such as obtained by `debugTrace` or `eth_call` RPC method.
```json
{
    "0x1b6bc8e955c946b4ceb8800f584828892f28b569c44c6b36d45f841c5ffb5cff": {
        "txhash": "0x1b6bc8e955c946b4ceb8800f584828892f28b569c44c6b36d45f841c5ffb5cff",
        "calls": [],
        "from": "0x77ab20c90882cd658c8f4d354394be4daccdbf15",
        "gas": "0x264c8",
        "gasUsed": "0x5a07",
        "input": "0xa9059cbb000000000000000000000000534ac36823b3f5b2e953029dcabbaa5af814e9690000000000000000000000000000000000000000000000000000005effd29bf6",
        "output": "0x0000000000000000000000000000000000000000000000000000000000000001",
        "to": "0x5e3002dff591c5e75bb9dedae268049742e6b13a",
        "type": "CALL",
        "value": "0x0",
        "logs": [
            {
                "index": 0,
                "address": "0x5e3002dff591c5e75bb9dedae268049742e6b13a",
                "topics": [
                    "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
                    "0x00000000000000000000000077ab20c90882cd658c8f4d354394be4daccdbf15",
                    "0x000000000000000000000000534ac36823b3f5b2e953029dcabbaa5af814e969"
                ],
                "data": "0x0000000000000000000000000000000000000000000000000000005effd29bf6"
            }
        ]
    }
}
```
As for the outputs, they are a bit complicated. key is still the transaction hash while the value is nodes in the call graph. the value in the list of nodes is maps from the address(EoA or CA) to edges list that the call involved by the node. 
From the perspective of ethereum, the log should store in a block, especially belonging to the address who emitted it. However, for the convenience of ML training, it stores in the `log` filed  of edge whose `to` is the address have the log.
If the edges of a node is null, it means no further call from this node because the edges demonstrate calls FROM the node.

In the `functionData`, if the call can be decoded by function signature or ABI databases, the `inputParams` or `outputParams`, both omittable if empty,  have the types and values of function arguments. Even the `functionData` is omittable while
can not decode the calldata. The decode of logs is quite time-consuming and laborious. the `parameters` is only available when the fully correct ABI for the `to` address of this call(or internal call) exists in ABI database. Otherwise, original logs 
is provided with  just `eventSignature` or without it if signature database can not match the event id, namely the topics[0]. Basically, we assume the hash conflict is impossible for event id.

```json
{
"0x003b945ee64109b239981c07c21648f18bdd7a0bcc037a03fb50981629fcd39e": {
    "nodes": {
      "0x6ff67737596e36d0194f225aee0e3519c5ad14cb": {
        "edges": [
          {
            "to": "0xc5b106f17246b2f5c0c658dbd6e8d168695806ab",
            "type": "CALL",
            "gas": "0x10538",
            "gasUsed": "0x920a",
            "input": "0xa9059cbb00000000000000000000000040b583b0bfce0fa28fa61539c9b8e012ecb7f9b0000000000000000000000000000000000000000000000000000000030263d940",
            "output": "0x0000000000000000000000000000000000000000000000000000000000000001",
            "value": "0x0",
            "functionData": {
              "name": "transfer",
              "inputParams": [
                {
                  "type": "address",
                  "value": "0x40B583b0BFCe0fA28fA61539c9b8E012eCb7f9b0"
                },
                {
                  "type": "uint256",
                  "value": "12925000000"
                }
              ],
              "outputParams": [
                {
                  "type": "address",
                  "value": "0x40B583b0BFCe0fA28fA61539c9b8E012eCb7f9b0"
                },
                {
                  "type": "uint256",
                  "value": "12925000000"
                },
                {
                  "type": "bool",
                  "value": "true"
                }
              ]
            },
            "logs": [
              {
                "index": 0,
                "eventSignature": "Transfer(address,address,uint256)",
                "parameters": [
                  {
                    "name": "_from",
                    "type": "address",
                    "value": "0x0000000000000000000000006ff67737596e36d0194f225aee0e3519c5ad14cb"
                  },
                  {
                    "name": "_to",
                    "type": "address",
                    "value": "0x00000000000000000000000040b583b0bfce0fa28fa61539c9b8e012ecb7f9b0"
                  },
                  {
                    "name": "_value",
                    "type": "uint256",
                    "value": "12925000000"
                  }
                ]
              }
            ]
          }
        ]
      },
      "0xc5b106f17246b2f5c0c658dbd6e8d168695806ab": {
        "edges": null
      }
    }
  }
}
```