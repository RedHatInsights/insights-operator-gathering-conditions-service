package service

import (
	"reflect"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
)

type ClusterMapping [][]string

func (cm ClusterMapping) IsValid() bool {
	// TODO: Check filepath exists
	// TODO: Add UTs
	// TODO: Add support for non v* versions

	versions := []string{} // used to check if it's sorted
	for _, slice := range cm {
		if len(slice) != 2 {
			log.Error().Int("len", len(slice)).Strs("slice", slice).Msg("Unexpected slice length")
			return false
		}
		version := slice[0]
		if !semver.IsValid(version) {
			log.Error().Str("version", version).Msg("Invalid semver")
			return false
		}
		versions = append(versions, version)
	}

	sortedVersions := make([]string, len(versions))
	copy(sortedVersions, versions)
	semver.Sort(sortedVersions)
	if !reflect.DeepEqual(sortedVersions, versions) {
		log.Error().Strs("sortedVersions", sortedVersions).Strs("versions", versions).Msg("Cluster mapping is not sorted")
		return false
	}

	return true
}
