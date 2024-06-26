package main

import (
	"path/filepath"
	abiDecoder "tx-analyze/abidec"
	"tx-analyze/txgraph"
)

func main() {
	fromDir := "./datum/txs/labeled"
	toDir := "./datum/txs/results"

	dbBaseDir := "/root/eth/eth-abi/datum/dbs"
	dbDir := make([]string, 3)
	for i := 0; i < 3; i++ {
		dbDir[0] = filepath.Join(dbBaseDir, "abi")
		dbDir[1] = filepath.Join(dbBaseDir, "func_sig")
		dbDir[2] = filepath.Join(dbBaseDir, "event_sig")
	}

	var d abiDecoder.ABIDB
	d.OpenDBs(dbDir)

	txgraph.MakeGraph(fromDir, toDir, &d, 20, 200, false)

}
