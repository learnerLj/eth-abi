package main

import "tx-analyze/txgraph"

func main() {
	fromDir := "./datum/txs/labeled"
	toDir := "./datum/txs/results"

	sigDb := "./datum/dbs/func_sig.json"
	eventDb := "./datum/dbs/events_sig.json"
	abiDb := "./datum/dbs/addr2abi.json"

	txgraph.MakeGraph(fromDir, toDir, sigDb, eventDb, abiDb, 20, 200, true)

}
