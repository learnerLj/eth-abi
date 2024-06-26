package txgraph

import (
	"sync"
	abiDecoder "tx-analyze/abidec"
	"tx-analyze/utils"

	"tx-analyze/types"

	"log"
	"tx-analyze/staticts"

	cmap "github.com/orcaman/concurrent-map/v2"
)

func parseCall(call *types.Call, graph *types.Graph, db *abiDecoder.ABIDB) {

	// create node if not exist
	fromNode, fromExists := graph.Nodes[call.From]
	if !fromExists {
		fromNode = &types.Node{}
	}
	_, toExists := graph.Nodes[call.To]
	if !toExists {
		graph.Nodes[call.To] = &types.Node{}
	}

	staticts.TotalCall += 1
	staticts.TotalEvent += len(call.CallLogs)
	if len(call.Input) >= 10 {
		staticts.TotalCallFunc++
	}

	//var ethABI *ethabi.ABI
	//var abiSource abiDecoder.ABISource

	var funcData *types.MethodData

	if abi, err := db.LoadABIFromAddr(call.To); err == nil {
		dec := abiDecoder.ABIDec{DecABI: abi}

		dec.ParamsForLogs(call.CallLogs)

		md, err := dec.DecodeInput(call.Input)
		if err != nil {
			log.Printf("%v, address: %v, hash: %v", err, call.To, call.Txhash)
			//if len(call.Txhash) != 0 {
			//	log.Printf("%v, address: %v, hash: %v", err, call.To, call.Txhash)
			//} else {
			//	log.Printf("%v, address: %v", err, call.To)
			//}
		} else { // decode parameters for return values
			staticts.ABIExistCallLegalParam++
			err := dec.DecodeOutput(md, call.Input, call.Output)
			if err != nil {
				log.Printf("%v. address: %v", err, call.To)
			}
			//else {
			//	staticts.ABIExistCallLegalOutput++
			//}
			funcData = md
		}

	} else if abis, err := db.LoadABIFromFuncSig(call.To); err == nil {
		staticts.FuncSigExistCall++

		for _, myabi := range abis {
			//decoder, err := abiDecoder.ABIByFuncSig(abi)
			dec := abiDecoder.ABIDec{DecABI: myabi}
			if err != nil {
				log.Printf("invalid func abi for external call。 %v", err)
				continue
			}
			if funcWithParams, err := dec.DecodeInput(call.Input); err == nil {
				staticts.FuncSIgParamCall++

				funcData = funcWithParams
				break
			}
			log.Printf("%s matchs one of %v but was not included now", call.Input[:10])
		}

		//TODO: decode the event
		//for _, logItem := range call.CallLogs {
		//
		//	if logItem == nil {
		//		log.Println(call.CallLogs)
		//		os.Exit(99)
		//	}
		//
		//	//默认不存在anonymous事件，或者说即使是anonymous，但是第一个参数不可能匹配上对应的 event selector.
		//	if logItem.Topics == nil || len(logItem.Topics) == 0 {
		//		continue
		//	}
		//	id := logItem.Topics[0]
		//
		//	myabi, err := db.LoadABIFromEventSig(id)
		//	if err != nil {
		//		fmt.Println(err)
		//		continue
		//	}
		//	db.DecABI = myabi
		//	if err := db.ParamsForSingleLog(logItem); err != nil {
		//		log.Println(err, "\ncontract address: ", logItem.Address)
		//	}
		//}
	} else {
		funcData = nil
	}

	//// if abi exists for this call
	//if jsonABIByAddr, exist := addrABIMap[call.To]; exist {
	//	staticts.ABIExistCall++
	//
	//	if myABI, err := abiDecoder.SetABI(jsonABIByAddr); err != nil { //ABI exists but not fit the address
	//		log.Printf("%v, for address %s", err, call.To)
	//
	//	} else {
	//		staticts.ABIExistCallLegal++
	//		//get parameters of logs
	//		myABI.ParamsForLogs(call.CallLogs)
	//
	//		//get parameters for function input
	//		if md, err := myABI.DecodeInput(call.Input); err != nil {
	//			if len(call.Txhash) != 0 {
	//				log.Printf("%v, address: %v, hash: %v", err, call.To, call.Txhash)
	//			} else {
	//				log.Printf("%v, address: %v", err, call.To)
	//			}
	//		} else { // decode parameters for return values
	//			staticts.ABIExistCallLegalParam++
	//			if err := myABI.DecodeOutput(md, call.Input, call.Output); err != nil {
	//				log.Printf("%v. address: %v", err, call.To)
	//			} else {
	//				staticts.ABIExistCallLegalOutput++
	//			}
	//			funcData = md
	//		}
	//
	//	}
	//
	//} else if sigs := call.Sigs(funcSigMap); sigs != nil { //only have function signature or event signature
	//	//get event signature
	//	staticts.FuncSigExistCall++
	//	for _, logItem := range call.CallLogs {
	//
	//		if logItem == nil {
	//			log.Println(call.CallLogs)
	//			os.Exit(99)
	//		}
	//
	//		//默认不存在anonymous事件，或者说即使是anonymous，但是第一个参数不可能匹配上对应的 event selector.
	//		if logItem.Topics == nil || len(logItem.Topics) == 0 {
	//			continue
	//		}
	//		id := logItem.Topics[0]
	//		if sig, exist := eventSigMap[id]; exist {
	//			logItem.EventSignature = sig
	//			staticts.EventByEventcSig++
	//		}
	//	}
	//
	//	//get parameters
	//	funcData = decodeInputWithPossibleFuncSignatures(sigs, call.Input)
	//} else {
	//	funcData = nil
	//}

	// create edge and add to the node which initiates it
	edge := utils.NewEdge(call)
	edge.FuncData = funcData

	fromNode.Edges = append(fromNode.Edges, edge)
	graph.Nodes[call.From] = fromNode

	//graph.Nodes[call.To].ContractLog = append(graph.Nodes[call.To].ContractLog, call.CallLogs...)

	for _, subCall := range call.Calls {
		if call != nil {
			parseCall(subCall, graph, db)
		}
	}
}

func decodeInputWithPossibleFuncSignatures(funcSigs []string, input string) *types.MethodData {
	if len(input) < 10 {
		return nil
	}
	for _, sig := range funcSigs {
		decoder, err := abiDecoder.ABIByFuncSig(sig)
		if err != nil {
			log.Printf("invalid func sig for external call。 %v", err)
			continue
		}
		if funcWithParams, err := decoder.DecodeInput(input); err == nil {
			staticts.FuncSIgParamCall++
			return funcWithParams
		}
	}
	log.Printf("%s matchs one of %v but was not included now", input[:10], funcSigs)
	return nil
}

func scanTx(txGraph *cmap.ConcurrentMap[string, types.Graph], c *types.Call, d *abiDecoder.ABIDB) {
	txhash := c.Txhash
	//init graph for each transaction
	graph := types.Graph{Nodes: make(map[string]*types.Node)}

	staticts.TotalTx++

	parseCall(c, &graph, d) // decode transaction trace

	txGraph.Set(txhash, graph)

}

func scanFile(path string, txGraph *cmap.ConcurrentMap[string, types.Graph], maxConcurrentTx int, db *abiDecoder.ABIDB) {
	// var txsCalls map[string]types.Call
	txsCalls := utils.LoadCalls(path) //parse JSON to Calls
	semaphore := make(chan struct{}, maxConcurrentTx)
	wg := sync.WaitGroup{}

	for _, c := range txsCalls {
		semaphore <- struct{}{}
		wg.Add(1)
		cnt := utils.CountCalls(&c)
		utils.SyncMapAdd1(&staticts.CallsInSingleTx, cnt)
		copyc := c // avoid potential concurrent bug
		go func(call *types.Call) {
			defer wg.Done()
			scanTx(txGraph, call, db)
			<-semaphore
		}(&copyc)
	}
	wg.Wait()

}
