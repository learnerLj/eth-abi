package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"sync"
	"tx-analyze/types"

	cmap "github.com/orcaman/concurrent-map/v2"
)

func NewEdge(call *types.Call) *types.Edge {
	return &types.Edge{
		To:           call.To,
		Type:         call.Type,
		Gas:          call.Gas,
		GasUsed:      call.GasUsed,
		Input:        call.Input,
		Output:       call.Output,
		Value:        call.Value,
		Error:        call.Error,
		RevertReason: call.RevertReason,
		ContractLog:  call.CallLogs,
	}
}

func LoadFunctionSignatures(path string) (funcSigMap map[string][]string) {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	err = decoder.Decode(&funcSigMap)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func LoadGraph(path string) (graph map[string]types.Graph) {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	err = decoder.Decode(&graph)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func LoadCalls(path string) (call map[string]types.Call) {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	err = decoder.Decode(&call)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func LoadEventSignatures(path string) (eventSigMap map[string]string) {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	err = decoder.Decode(&eventSigMap)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func LoadABIs(filename string) (addrABIMap map[string]string) {
	addrABIMap = make(map[string]string)
	// 读取 JSON 文件
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&addrABIMap)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("ABI database loaded!")
	return
}

func CountCalls(c *types.Call) int {
	cnt := 1
	if c.Calls == nil {
		return cnt
	}

	for _, subCall := range c.Calls {
		cnt += CountCalls(subCall)
	}
	return cnt
}

func GraphToFile(jsonMap *cmap.ConcurrentMap[string, types.Graph], toPath string) error {
	fileLog, err := os.OpenFile(toPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error opening file:", err)
		return err
	}
	defer fileLog.Close()

	encoder := json.NewEncoder(fileLog)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonMap); err != nil {
		log.Fatal("Failed to write graph to file:", err)
		return err
	}

	return nil
}

func ABIToFile(jsonMap map[string]string, toPath string) error {

	fileLog, err := os.OpenFile(toPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error opening file:", err)
		return err
	}
	defer fileLog.Close()

	encoder := json.NewEncoder(fileLog)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonMap); err != nil {
		log.Fatal("Failed to write to file:", err)
		return err
	}

	log.Printf("ABI written to %v", toPath)

	return nil
}

func Init() {

}

func unescapeString(input string) string {
	// 替换转义的引号，但保留 "" 为空字符串
	re := regexp.MustCompile(`\\n|\\\\`)
	output := re.ReplaceAllStringFunc(input, func(match string) string {
		switch match {
		case "\\n":
			return "\n"
		case "\\\\":
			return "\\"
		}
		return match
	})

	// 替换转义的双引号
	re = regexp.MustCompile(`\\"`)
	output = re.ReplaceAllString(output, `"`)

	return output
}

func SyncMapAdd1(m *sync.Map, key int) {
	if value, ok := m.Load(key); ok {
		intValue, _ := value.(int)
		m.Store(key, intValue+1)
	} else {
		m.Store(key, 1)
	}
}

func ToGobBytes(in interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(in)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FromGobBytes(out interface{}, data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(out)
	if err != nil {
		return err
	}
	return nil
}
