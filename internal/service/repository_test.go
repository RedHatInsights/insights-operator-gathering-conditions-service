package service_test

import (
	"testing"

	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	type testCase struct {
		name            string
		mockData        []byte
		expectedAnError bool
		expectedRules   service.Rules
	}

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
			mockData:        []byte(validRulesJSON),
			expectedAnError: false,
			expectedRules:   validRules,
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
