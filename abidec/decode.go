package abiDecoder

import (
	"encoding/hex"
	"encoding/json"
	"github.com/dgraph-io/badger"
	"strings"
	"tx-analyze/staticts"
	"tx-analyze/types"

	"fmt"
	"log"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func SetABI(contractAbi string) (*ABIDecoder, error) {
	myabi, err := ethabi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		return nil, fmt.Errorf("invalid ABI, %v", err)
	}
	return &ABIDecoder{DecABI: &myabi}, nil
}

type ABISource uint8

const (
	FromAddr ABISource = iota
	FromFuncSig
	FromEventSig
)

type ABIDecoder struct {
	DecABI     *ethabi.ABI
	ABIDb      *badger.DB
	FuncSigDb  *badger.DB
	EventSigDb *badger.DB
}

func (d *ABIDecoder) DecodeInput(txInput string) (*types.MethodData, error) {
	//no function called
	if len(txInput) < 10 {
		return nil, fmt.Errorf("not call any function because input is too short")
	}
	txInput = strings.TrimPrefix(txInput, "0x") // skip 0x prefix
	decodedSig, err := hex.DecodeString(txInput[:8])
	if err != nil {
		return nil, err
	}

	method, err := d.DecABI.MethodById(decodedSig)
	if err != nil {
		return nil, err
	}

	decodedData, err := hex.DecodeString(txInput[8:])
	if err != nil {
		return nil, err
	}

	inputs, err := method.Inputs.Unpack(decodedData)
	if err != nil {
		return nil, err
	}

	retData := types.MethodData{Name: method.Name}
	nonIndexedArgs := method.Inputs.NonIndexed()

	for i, input := range inputs {
		arg := nonIndexedArgs[i]
		param := types.ParamData{
			Value: fmt.Sprintf("%v", input),
			Type:  arg.Type.String(),
		}
		retData.InputParams = append(retData.InputParams, param)
	}
	return &retData, nil
}

func (d *ABIDecoder) DecodeOutput(funcData *types.MethodData, funcSig, txOutput string) error {
	funcSig = strings.TrimPrefix(funcSig, "0x")
	//no function called
	if len(funcSig) < 8 {
		return nil
	}

	decodedSig, err := hex.DecodeString(funcSig[:8])
	if err != nil {
		return err
	}

	method, err := d.DecABI.MethodById(decodedSig)
	if err != nil {
		return err
	}

	txOutput = strings.TrimPrefix(txOutput, "0x")
	if len(txOutput) > 0 {
		decodedOutput, _ := hex.DecodeString(txOutput)

		if outputs, err := method.Outputs.Unpack(decodedOutput); err != nil {
			return fmt.Errorf("ABI does not match the call")
		} else {
			outputsArgs := method.Outputs.NonIndexed()
			for i, output := range outputs {
				arg := outputsArgs[i]
				param := types.ParamData{
					Value: fmt.Sprintf("%v", output),
					Type:  arg.Type.String(),
				}
				funcData.OutputParams = append(funcData.InputParams, param)
			}
		}

	}

	return nil
}

func (d *ABIDecoder) getParamsForSingleLog(logItem *types.SingleLog) error {
	if len(logItem.Topics) == 0 {
		return fmt.Errorf("anonymous event has no signature")
	}
	var eventId [32]byte
	idHex := logItem.Topics[0]
	idBytes, err := hex.DecodeString(idHex[2:])
	if err != nil {
		log.Fatal(err)
	}
	if len(idBytes) == 32 {
		copy(eventId[:], idBytes)
	} else {
		log.Fatalf("event id should be 32 bytes. Event id(Topic[0]): %s", idHex)
	}

	event, err := d.DecABI.EventByID(common.Hash(idBytes))
	if err != nil {
		printEvent := make(map[string]struct {
			Sig string `json:"Sig"`
			ID  string `json:"ID"`
		})

		for k, e := range d.DecABI.Events {
			// 创建一个新的结构体实例
			event := struct {
				Sig string `json:"Sig"`
				ID  string `json:"ID"`
			}{
				ID:  e.ID.Hex(),
				Sig: e.Sig,
			}
			printEvent[k] = event
		}
		b, _ := json.MarshalIndent(printEvent, "", "  ")
		return fmt.Errorf("no such event in ABI. Provided:\n%s\nbut meet event id: %s", string(b), idHex)
	}

	data, err := hex.DecodeString(logItem.Data[2:])
	if err != nil {
		log.Fatal(err)
	}
	dataList, err := d.DecABI.Unpack(event.Name, data)

	if err != nil {
		return fmt.Errorf("%v\nEvent %s whoes id is %s with data %s", err, event.Sig, idHex, logItem.Data)
	}
	// if the number of indexed parameters and the number of non-indexed parameters do not sum up to the number of
	// parameters of the event in ABI, the called contract does not match the abi.
	if len(logItem.Topics)-1 != len(event.Inputs)-len(dataList) {
		return fmt.Errorf("indexed parameters with in log is %v but expect %v in ABI", len(logItem.Topics)-1, len(event.Inputs)-len(dataList))
	}

	params := make([]*types.ParamData, 0, len(event.Inputs))
	topicIndex := 1 //indexed value are put in topic
	dataIndex := 0  // no indexed value are put in data
	for _, input := range event.Inputs {
		param := &types.ParamData{Name: input.Name, Type: input.Type.String()}

		var value interface{}
		if input.Indexed {
			value = logItem.Topics[topicIndex]
			topicIndex++
		} else {
			value = dataList[dataIndex]
			dataIndex++
		}
		param.Value = fmt.Sprintf("%v", value)

		params = append(params, param)
	}

	logItem.Params = params
	logItem.EventSignature = event.Sig

	return nil
}

func (d *ABIDecoder) ParamsForLogs(logItems []*types.SingleLog) {
	for _, logItem := range logItems {
		// staticts.TotalEvent++

		//If it cannot be resolved, it will remain intact because there may be an ABI defect.
		//Otherwise, parameters and signatures are parsed
		if err := d.getParamsForSingleLog(logItem); err != nil {
			log.Println(err, "\ncontract address: ", logItem.Address)
			staticts.ABIExistCallLegalEventFailed++
		}
	}

}
