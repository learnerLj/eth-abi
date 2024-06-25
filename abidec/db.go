package abiDecoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"log"
	"strings"
	"tx-analyze/utils"
)

var (
	errABINotFound = errors.New("ABI not find for this address")
	errABIFormat   = errors.New("invalid ABI")

	errFuncSigNotFound = errors.New("function signature not found for this selector")
	errFuncSigFormat   = errors.New("function signature wrong format")

	errEventNotFound = errors.New("event signature not found")
)

func (d *ABIDecoder) LoadABIFromAddr(addr string) (*ethabi.ABI, error) {
	abiStr, err := d.ReadABI(addr)
	if err != nil {
		return nil, errABINotFound
	}
	myabi, err := ethabi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, errABIFormat
	}
	return &myabi, nil
}

func (d *ABIDecoder) LoadABIFromFuncSig(funcSelector string) ([]*ethabi.ABI, error) {
	funcSigStrs, err := d.ReadFuncSig(funcSelector)
	if err != nil {
		return nil, errFuncSigNotFound
	}

	abis := make([]*ethabi.ABI, 0, len(funcSigStrs))
	for i, funcSigStr := range funcSigStrs {
		fs, err := ParseFunctionSignature(funcSigStr)
		if err != nil {
			return nil, errFuncSigFormat
		}
		jsonABI, _ := json.Marshal(fs)

		abiStr := string(jsonABI)
		myabi, err := ethabi.JSON(strings.NewReader(abiStr))
		if err != nil {
			return nil, errFuncSigFormat
		}
		abis[i] = &myabi
	}

	return abis, nil
}

func (d *ABIDecoder) LoadABIFromEventSig(eventHash string) (*ethabi.ABI, error) {
	eventSig, err := d.ReadEventSig(eventHash)
	if err != nil {
		return nil, errEventNotFound
	}
	fmt.Println(eventSig)
	//TODO: convert event signature to json-ABI
	return nil, nil

}

// abi, funcsig, eventsig
func (d *ABIDecoder) OpenDBs(dbPath []string) {
	for i, path := range dbPath {
		db, err := badger.Open(badger.DefaultOptions(path))
		if err != nil {
			log.Fatal(err)
		}
		switch i {
		case 0:
			d.ABIDb = db
		case 1:
			d.FuncSigDb = db
		case 2:
			d.EventSigDb = db
		default:
			log.Fatal("unexpected db paths")
		}
	}
}

func (d *ABIDecoder) CloseDBs() {
	d.ABIDb.Close()
	d.FuncSigDb.Close()
	d.EventSigDb.Close()
}

func (d *ABIDecoder) ReadABI(addr string) (string, error) {
	b, err := d.badgerRead(d.ABIDb, addr)
	if err != nil {
		return "", err
	}
	var abiStr string

	err = utils.FromGobBytes(&abiStr, b)
	if err != nil {
		return "", err
	}

	return abiStr, nil
}

func (d *ABIDecoder) WriteABI(key, value string) error {
	b, err := utils.ToGobBytes(value)
	if err != nil {
		return err
	}
	return d.badgerWrite(d.ABIDb, key, b)
}

func (d *ABIDecoder) ReadFuncSig(key string) ([]string, error) {
	b, err := d.badgerRead(d.FuncSigDb, key)
	if err != nil {
		return nil, err
	}
	var funcSigStr []string

	err = utils.FromGobBytes(&funcSigStr, b)
	if err != nil {
		return nil, err
	}

	return funcSigStr, nil
}

func (d *ABIDecoder) WriteFuncSig(key string, value []string) error {
	b, err := utils.ToGobBytes(value)
	if err != nil {
		return err
	}
	return d.badgerWrite(d.FuncSigDb, key, b)
}

func (d *ABIDecoder) ReadEventSig(key string) (string, error) {
	b, err := d.badgerRead(d.EventSigDb, key)
	if err != nil {
		return "", err
	}
	var eventSigStr string

	err = utils.FromGobBytes(&eventSigStr, b)
	if err != nil {
		return "", err
	}
	return eventSigStr, nil
}

func (d *ABIDecoder) WriteEventSig(key, value string) error {
	b, err := utils.ToGobBytes(value)
	if err != nil {
		return err
	}
	return d.badgerWrite(d.EventSigDb, key, b)
}

func (d *ABIDecoder) badgerRead(db *badger.DB, key string) ([]byte, error) {
	var value []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(v []byte) error {
			value = v
			return nil
		})
		return err
	})
	return value, err
}

func (d *ABIDecoder) badgerWrite(db *badger.DB, key string, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}
