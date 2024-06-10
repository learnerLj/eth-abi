package abiDecoder

import (
	"fmt"
	"github.com/dgraph-io/badger"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"log"
	"strings"
	"tx-analyze/utils"
)

func (d *ABIDecoder) LoadABIFromAddr(abiStr string) error {
	myabi, err := ethabi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return fmt.Errorf("invalid ABI, %v", err)
	}
	d.DecABI = &myabi
	return nil
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

func (d *ABIDecoder) ReadABI(key string) (string, error) {
	b, err := d.badgerRead(d.ABIDb, key)
	if err != nil {
		return "", err
	}
	var abiStr string

	err = utils.FromGobBytes(abiStr, b)
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

	err = utils.FromGobBytes(funcSigStr, b)
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

	err = utils.FromGobBytes(eventSigStr, b)
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
