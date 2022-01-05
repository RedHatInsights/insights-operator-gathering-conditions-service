package service_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	type testCase struct {
		name            string
		mockData        []byte
		expectedAnError bool
	}

	testCases := []testCase{
		{
			name:            "valid rule stored",
			mockData:        []byte(validRulesJSON),
			expectedAnError: false,
		},
		{
			name:            "invalid rule stored",
			mockData:        []byte("not a rule"),
			expectedAnError: true,
		},
	}

	for _, tc := range testCases {

		store := mockStorage{
			mockData: tc.mockData,
		}
		repo := service.NewRepository(&store)
		svc := service.New(repo)

		// Create the request:
		req, err := http.NewRequest("GET", "/gathering_rules", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder() // Used to record the response.
		handler := service.NewHandler(svc)

		router := mux.NewRouter()

		handler.Register(router)

		router.ServeHTTP(rr, req)

		if tc.expectedAnError {
			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			assert.Contains(t, rr.Body.String(), "error")
		} else {
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Contains(
				t,
				rr.Body.String(),
				`"rules":[{"conditions":["condition 1","condition 2"],"gathering_functions":"the gathering functions"}]`)
		}
	}

}
