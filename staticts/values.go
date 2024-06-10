package staticts

import (
	"fmt"
	"sort"
	"sync"
)

// all calls involve function
var TotalTx int
var TotalCall int //all calls
var TotalCallFunc int
var ABIExistCall int      // decoding with abi on contract address
var ABIExistCallLegal int // abi find in database but not match this call data

// event
var TotalEvent int
var ABIExistCallLegalEventFailed int
var EventByEventcSig int

// function data
var FuncSigExistCall int
var FuncSIgParamCall int

// var FuncSIgOutputCall int

var ABIExistCallLegalParam int
var ABIExistCallLegalOutput int

// ---
// TODO: use github.com/dgraph-io/badger to replace the map for the use of fast key-value (KV) database
var CallsInSingleTx sync.Map

type CallNum struct {
	calls   int
	counter int
}

func callsInTx() {

	var nums []CallNum

	CallsInSingleTx.Range(func(key, value interface{}) bool {
		keyInt, _ := key.(int)
		valueInt, _ := value.(int)
		calls, count := keyInt, valueInt
		nums = append(nums, CallNum{calls, count})
		return true
	})

	// 使用sort.Slice对nums进行排序，基于calls字段
	sort.Slice(nums, func(i, j int) bool {
		return nums[i].calls < nums[j].calls
	})

	length := len(nums)
	maxnNum := nums[length-1]
	percentile50 := 0
	percentile80 := 0
	percentile95 := 0

	tmpCnt := 0
	for _, v := range nums {
		tmpCnt += v.counter
		if tmpCnt*100/TotalTx > 50 && percentile50 == 0 {
			percentile50 = v.calls
		} else if tmpCnt*100/TotalTx > 80 && percentile80 == 0 {
			percentile80 = v.calls
		} else if tmpCnt*100/TotalTx > 95 && percentile95 == 0 {
			percentile95 = v.calls
		}

	}

	fmt.Printf("max calls in single Tx is: %v\npercentile50,80,95=%v, %v, %v\n", maxnNum.calls, percentile50, percentile80, percentile95)

	if len(nums) > 3 {
		first := nums[0]
		second := nums[1]
		third := nums[2]
		r1 := float64(first.counter*100) / float64(TotalTx)
		r2 := r1 + float64(second.counter*100)/float64(TotalTx)
		r3 := r2 + float64(third.counter*100)/float64(TotalTx)
		fmt.Printf("Txs within %v,%v,%v calls account for %.2f%%,%.2f%%,%.2f%%\n", first.calls, second.calls, third.calls,
			r1, r2, r3)
	}
}

func Statistic() {
	callsPerTx := float64(TotalCall) / float64(TotalTx)
	callsInvokeFuncRate := 100 * float64(TotalCallFunc) / float64(TotalCall)
	ABIRate := float64(ABIExistCall*100) / float64(TotalCallFunc)
	ABICorrectRate := float64(ABIExistCallLegal*100) / float64(ABIExistCall)
	eventParamABIRate := 100 - float64(ABIExistCallLegalEventFailed*100)/float64(TotalEvent)
	EventByEventcSigRate := float64(EventByEventcSig*100) / float64(TotalEvent)
	funcDataABIRate := float64(ABIExistCallLegalParam*100) / float64(TotalCallFunc)
	funcOutputRate := float64(ABIExistCallLegalOutput*100) / float64(TotalCallFunc)
	funcSIgParamCallRate := float64(FuncSIgParamCall*100) / float64(TotalCallFunc)

	fmt.Printf("processed transactions: %v in which contain calls: %v of which invokes function:  %.2f%%, average calls per tx: %.2f\n", TotalTx, TotalCall, callsInvokeFuncRate, callsPerTx)
	callsInTx()
	fmt.Printf("calls that find ABI via address: %.2f%%, in which format-correct ones have %.2f%%\n", ABIRate, ABICorrectRate)
	fmt.Printf("Param by ABI: %.2f%%, by signature %.2f%%\n", funcDataABIRate, funcSIgParamCallRate)
	fmt.Printf("Output by ABI: %.2f%%\n", funcOutputRate)
	fmt.Printf("Total events: %v, that decoding parameters if ABI exists: %.2f%%, only get event signature: %.2f%%\n", TotalEvent, eventParamABIRate, EventByEventcSigRate)
}
