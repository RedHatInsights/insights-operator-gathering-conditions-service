package service_test

import (
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testFilesPath = "../../tests/rapid-recommendations"

func TestClusterMappingIsValid(t *testing.T) {
	t.Run("valid map", func(t *testing.T) {
		var sut service.ClusterMapping = [][]string{
			{"1.0.0", "experimental_1.json"},
			{"2.0.0", "experimental_2.json"},
			{"3.0.0", "config_default.json"},
		}
		assert.True(t, sut.IsValid(testFilesPath))
	})

	t.Run("invalid map: invalid version", func(t *testing.T) {
		var sut service.ClusterMapping = [][]string{
			{"1.0.0", "experimental_1.json"},
			{"not a valid version", "experimental_2.json"},
			{"3.0.0", "config_default.json"},
		}
		assert.False(t, sut.IsValid(testFilesPath))
	})

	t.Run("invalid map: JSON not found", func(t *testing.T) {
		var sut service.ClusterMapping = [][]string{
			{"1.0.0", "experimental_1.json"},
			{"2.0.0", "not-found.json"},
			{"3.0.0", "config_default.json"},
		}
		assert.False(t, sut.IsValid(testFilesPath))
	})
}

func TestClusterMappingGetFilepathForVersion(t *testing.T) {
	t.Run("valid map", func(t *testing.T) {
		type testCase struct {
			ocpVersion   string
			wantFilepath string
			expectError  bool
		}
		testCases := []testCase{
			{
				ocpVersion:   "0.1.0",
				expectError:  true,
				wantFilepath: "",
			},
			{
				ocpVersion:   "1.0.0",
				wantFilepath: "experimental_1.json",
			},
			{
				ocpVersion:   "1.5.0",
				wantFilepath: "experimental_1.json",
			},
			{
				ocpVersion:   "2.0.0",
				wantFilepath: "experimental_2.json",
			},
			{
				ocpVersion:   "2.5.0",
				wantFilepath: "experimental_2.json",
			},
			{
				ocpVersion:   "3.0.0",
				wantFilepath: "config_default.json",
			},
			{
				ocpVersion:   "3.5.0",
				wantFilepath: "config_default.json",
			},
		}

		var sut service.ClusterMapping = [][]string{
			{"1.0.0", "experimental_1.json"},
			{"2.0.0", "experimental_2.json"},
			{"3.0.0", "config_default.json"},
		}

		for _, tc := range testCases {
			t.Run(tc.ocpVersion, func(t *testing.T) {
				ocpVersionParsed, _ := semver.Make(tc.ocpVersion)
				got, err := sut.GetFilepathForVersion(ocpVersionParsed)
				if tc.expectError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Equal(
					t,
					tc.wantFilepath,
					got)
			})
		}
	})
}
