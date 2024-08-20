package service

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/blang/semver/v4"
	"github.com/rs/zerolog/log"
)

type ClusterMapping [][]string

// IsValid check the list is in order (based on the versions), that the versions
// can be parsed and that the remote configurations are accessible
func (cm ClusterMapping) IsValid(remoteConfigurationPath string) bool {
	// TODO: Add UTs

	versions := []semver.Version{} // used to check if it's sorted
	for _, slice := range cm {
		if len(slice) != 2 {
			log.Error().Int("len", len(slice)).Strs("slice", slice).Msg("Unexpected slice length")
			return false
		}
		version := slice[0]
		versionParsed, err := semver.Make(version)
		if err != nil {
			log.Error().Str("version", version).Msg("Invalid semver")
			return false
		}
		versions = append(versions, versionParsed)

		filepath := slice[1]
		fullFilepath := fmt.Sprintf("%s/%s", remoteConfigurationPath, filepath)
		if _, err := os.Stat(fullFilepath); errors.Is(err, os.ErrNotExist) {
			log.Error().Str("filepath", fullFilepath).
				Msg("Remote configuration filepath couldn't be accessed")
			return false
		}
	}

	sortedVersions := make([]semver.Version, len(versions))
	copy(sortedVersions, versions)
	semver.Sort(sortedVersions)
	if !reflect.DeepEqual(sortedVersions, versions) {
		log.Error().Interface("sortedVersions", sortedVersions).
			Interface("versions", versions).
			Msg("Cluster mapping is not sorted")
		return false
	}

	return true
}
