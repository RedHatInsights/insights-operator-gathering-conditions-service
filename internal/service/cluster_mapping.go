package service

import (
	"reflect"

	"github.com/blang/semver/v4"
	"github.com/rs/zerolog/log"
)

type ClusterMapping [][]string

func (cm ClusterMapping) IsValid() bool {
	// TODO: Check filepath exists
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
