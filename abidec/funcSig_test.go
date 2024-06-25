package abiDecoder_test

import (
	"encoding/json"
	"errors"
	"fmt"
	badger "github.com/dgraph-io/badger"
	"log"
	"sync"
	"testing"
	"tx-analyze/abidec"
	"tx-analyze/utils"
)

const (
	badgerABI   = "../datum/dbs/abi"
	badgersig   = "../datum/dbs/func_sig"
	badgerEvent = "../datum/dbs/event_sig"

	sigDb   = "./datum/dbs/func_sig.json"
	eventDb = "./datum/dbs/events_sig.json"
	abiDb   = "./datum/dbs/addr2abi.json"
)

func TestFuncSigDecoder(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	sig := "doSomething(address,(address),(uint8,uint256),bytes)"
	data := "0xfd4ca6c00000000000000000000000005b38da6a701c568545dcfcb03fcb875f56beddc40000000000000000000000005b38da6a701c568545dcfcb03fcb875f56beddc400000000000000000000000000000000000000000000000000000000000000f000000000000000000000000000000000000000000000000000000000000026f800000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000049afd642100000000000000000000000000000000000000000000000000000000"

	d, err := abiDecoder.ABIByFuncSig(sig)
	if err != nil {
		log.Println(err)
	}
	p, err := d.DecodeInput(data)
	jsonP, _ := json.Marshal(p)
	fmt.Println(string(jsonP))
}

func TestJson2Badger(t *testing.T) {

	funcSigMap := utils.LoadFunctionSignatures(sigDb)
	eventSigMap := utils.LoadEventSignatures(eventDb)
	addrABIMap := utils.LoadABIs(abiDb)

	var d abiDecoder.ABIDecoder
	d.OpenDBs([]string{badgerABI, badgersig, badgerEvent})

	var sem = make(chan struct{}, 200)
	var wg sync.WaitGroup

	for k, v := range funcSigMap {
		sem <- struct{}{}
		wg.Add(1)
		go func(k string, v []string) {
			defer wg.Done()
			if _, err := d.ReadFuncSig(k); errors.Is(err, badger.ErrKeyNotFound) {

				if err := d.WriteFuncSig(k, v); err != nil {
					log.Fatal("badger write error!")
				}
				fmt.Println(k)
			}
			<-sem

		}(k, v)

	}

	for k, v := range eventSigMap {
		sem <- struct{}{}
		wg.Add(1)
		go func(k string, v string) {
			defer wg.Done()
			if _, err := d.ReadEventSig(k); errors.Is(err, badger.ErrKeyNotFound) {

				if err := d.WriteEventSig(k, v); err != nil {
					log.Fatal("badger write error!")
				}
				fmt.Println(k)
			}
			<-sem

		}(k, v)

	}

	for k, v := range addrABIMap {
		sem <- struct{}{}
		wg.Add(1)
		go func(k string, v string) {
			defer wg.Done()
			if _, err := d.ReadABI(k); errors.Is(err, badger.ErrKeyNotFound) {

				if err := d.WriteABI(k, v); err != nil {
					log.Fatal("badger write error!")
				}
				fmt.Println(k)
			}
			<-sem

		}(k, v)

	}

	wg.Wait()

	fmt.Println("finished")
}

func TestEventSIg(t *testing.T) {
	var d abiDecoder.ABIDecoder
	d.OpenDBs([]string{badgerABI, badgersig, badgerEvent})

	eventHash := "0x0000000008eb2ce0f0e5bde6b515488cb50d59e65e32f4d9f9d46a288dc72423"
	eventSig, _ := d.ReadEventSig(eventHash)
	fmt.Println(eventSig)

	funcSel := "0x00005b67"
	funcSig, _ := d.ReadFuncSig(funcSel)
	fmt.Println(funcSig)

	addr := "0x00000000000015bF55A34241Bbf73Ec4f4b080B2"
	abi, _ := d.ReadABI(addr)
	fmt.Println(abi)
}
