package types

import (
	"encoding/json"
	"os"
)

type Call struct {
	Txhash       string       `json:"txhash,omitempty"`
	Calls        []*Call      `json:"calls"`
	From         string       `json:"from"`
	To           string       `json:"to"`
	Type         string       `json:"type"`
	Gas          string       `json:"gas"`
	GasUsed      string       `json:"gasUsed"`
	Input        string       `json:"input"`
	Output       string       `json:"output"`
	Value        string       `json:"value,omitempty"`
	Error        string       `json:"error,omitempty"`
	RevertReason string       `json:"revertReason,omitempty"`
	CallLogs     []*SingleLog `json:"logs,omitempty"`
}

type Edge struct {
	To           string       `json:"to"`
	Type         string       `json:"type"`
	Gas          string       `json:"gas"`
	GasUsed      string       `json:"gasUsed"`
	Input        string       `json:"input"`
	Output       string       `json:"output"`
	Value        string       `json:"value,omitempty"`
	Error        string       `json:"error,omitempty"`
	RevertReason string       `json:"revertReason,omitempty"`
	FuncData     *MethodData  `json:"functionData,omitempty"`
	ContractLog  []*SingleLog `json:"logs,omitempty"`
}

type Node struct {
	// EAorCA      string             `json:"address"`
	Edges []*Edge `json:"edges"`
	//ContractLog []*SingleLog `json:"logs,omitempty"`
}

type Graph struct {
	Nodes map[string]*Node `json:"nodes"`
}

type SingleLog struct {
	Index          int          `json:"index"`
	Address        string       `json:"address"`
	Topics         []string     `json:"topics"`
	Data           string       `json:"data"`
	EventSignature string       `json:"eventSignature"`
	Params         []*ParamData `json:"parameters"`
}

type ParamData struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type"`
	Value string `json:"value"`
}
type MethodData struct {
	Name         string      `json:"name"`
	InputParams  []ParamData `json:"inputParams"`
	OutputParams []ParamData `json:"outputParams"`
}

func (s SingleLog) MarshalJSON() ([]byte, error) {
	// 定义一个辅助结构体来控制 JSON 序列化
	type Alias SingleLog

	if s.Params != nil {
		// 如果 Params 不为 nil，则不包含 Data 和 Topics 字段
		return json.Marshal(&struct {
			*Alias
			Address string   `json:"address,omitempty"`
			Data    string   `json:"data,omitempty"`
			Topics  []string `json:"topics,omitempty"`
		}{
			Alias: (*Alias)(&s),
		})
	} else {
		// 如果 Params 为 nil，则正常序列化，但不包含 Address 字段
		return json.Marshal(&struct {
			*Alias
			Address string `json:"address,omitempty"`
		}{
			Alias: (*Alias)(&s),
		})
	}
}

func (c Call) Sigs(funcSigMap map[string][]string) []string {
	if len(c.Input) >= 10 {
		bytesSig := c.Input[:10]

		if funcSigMap == nil {
			os.Exit(98)
		}

		if signatures, exist := funcSigMap[bytesSig]; exist {
			return signatures
		}
	}
	return nil
}
