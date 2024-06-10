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

func TestFuncSigDecoder(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	sig := "doSomething(address,(address),(uint8,uint256),bytes)"
	//0x5B38Da6a701c568545dCfcB03FcB875f56beddC4,[0x5B38Da6a701c568545dCfcB03FcB875f56beddC4],[240,9976],0x9afd6421
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
	sigDb := "./datum/dbs/func_sig.json"
	eventDb := "./datum/dbs/events_sig.json"
	abiDb := "./datum/dbs/addr2abi.json"

	badgerABI := "./datum/dbs/abi"
	badgersig := "./datum/dbs/func_sig"
	badgerEvent := "./datum/dbs/event_sig"

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
