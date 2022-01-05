package service_test

import "github.com/redhatinsights/insights-operator-conditional-gathering/internal/service"

var (
	validRules = service.Rules{
		Items: []service.Rule{
			{
				Conditions: []interface{}{
					"condition 1",
					"condition 2",
				},
				GatheringFunctions: "the gathering functions",
			},
		},
	}
	validRulesJSON = `
	{
		"rules": [
			{
				"conditions": ["condition 1", "condition 2"],
				"gathering_functions": "the gathering functions"
			}
		]
	}
	`
)

type mockStorage struct {
	mockData []byte
}

func (m *mockStorage) Find(string) []byte {
	return m.mockData
}
