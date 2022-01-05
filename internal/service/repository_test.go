package service_test

import (
	"testing"

	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	mockData []byte
}

func (m *mockStorage) Find(string) []byte {
	return m.mockData
}

type testCase struct {
	name            string
	mockData        []byte
	expectedAnError bool
	expectedRules   service.Rules
}

func TestRepository(t *testing.T) {
	validRules := `
	{
		"rules": [
			{
				"conditions": ["condition 1", "condition 2"],
				"gathering_functions": "the gathering functions"
			}
		]
	}
	`
	testCases := []testCase{
		{
			name:            "unparsable rule",
			mockData:        []byte("not a JSON"),
			expectedAnError: true,
		},
		{
			name:            "no data returned by storage",
			mockData:        nil,
			expectedAnError: true,
		},
		{
			name:            "valid rules returned by storage",
			mockData:        []byte(validRules),
			expectedAnError: false,
			expectedRules: service.Rules{
				Items: []service.Rule{
					{
						Conditions: []interface{}{
							"condition 1",
							"condition 2",
						},
						GatheringFunctions: "the gathering functions",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockStorage{
				mockData: tc.mockData,
			}
			r := service.NewRepository(&m)
			rules, err := r.Rules()
			if tc.expectedAnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tc.expectedRules, rules)
			}
		})
	}
}
