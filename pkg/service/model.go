package service

type GatheringFunctions struct {
}

type Conditions struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

type Rule struct {
	Conditions         []Conditions `json:"conditions,omitempty"`
	GatheringFunctions interface{}  `json:"gathering_functions,omitempty"`
}

type Rules struct {
	Items []Rule `json:"rules,omitempty"`
}
