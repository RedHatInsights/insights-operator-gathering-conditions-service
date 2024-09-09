package service

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	merrors "github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/errors"
	"github.com/blang/semver/v4"
	"github.com/rs/zerolog/log"
)

// ClusterMapping map OCP versions to remote configurations
type ClusterMapping [][]string

// IsValid check the list is in order (based on the versions), that the versions
// can be parsed and that the remote configurations are accessible
func (cm ClusterMapping) IsValid(remoteConfigurationPath string, conditionsVersion string) bool {
	versions := []semver.Version{} // used to check if it's sorted

	if len(cm) == 0 {
		log.Error().Interface("raw", cm).Msg("Cluster map needs to contain at least one pair of version and filepath")
		return false
	}

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
		fullFilepath := fmt.Sprintf("%s/%s/%s", remoteConfigurationPath, conditionsVersion, filepath)
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

// GetFilepathForVersion iterates over the cluster map returning the first
// filepath corresponding to the ocp version. Example:
/*
[
	["1.0.0", "first.json"],
	["2.0.0", "second.json"],
	["3.0.0", "third.json"]
] */
// would return first.json for versions between 1.0.0 and 2.0.0, second.json
// for versions between 2.0.0 and 3.0.0 and third.json for versions greater
// than 3.0.0
func (cm ClusterMapping) GetFilepathForVersion(ocpVersionParsed semver.Version) (string, error) {
	// check the version is not greater than the first slice
	firstVersion, err := semver.Make(cm[0][0])
	if err != nil {
		log.Error().Str("version", firstVersion.String()).Err(err).Msg("Invalid semver")
		return "", err
	}

	comparison := ocpVersionParsed.Compare(firstVersion)
	if comparison < 0 {
		errMsg := "the given OCP version is lower than the first one in the cluster map"
		log.Error().
			Str("version", firstVersion.String()).
			Str("ocpVersion", ocpVersionParsed.String()).
			Msg(errMsg)
		return "", &merrors.NotFoundError{
			ErrString: errMsg}
	} else if comparison == 0 {
		return cm[0][1], nil
	}

	previousFilepath := cm[0][1]
	for _, slice := range cm[1:] {
		version := slice[0]
		filepath := slice[1]
		versionParsed, err := semver.Make(version)

		if err != nil {
			log.Error().Str("version", version).Err(err).Msg("Invalid semver")
			return "", err
		}
		comparison := ocpVersionParsed.Compare(versionParsed)
		if comparison == 0 {
			// this means the ocp version is equal to the current version
			return filepath, nil
		} else if comparison < 0 {
			// this means the ocp version is below the current version
			return previousFilepath, nil
		}

		previousFilepath = filepath
	}

	log.Debug().Str("ocpVersion", ocpVersionParsed.String()).
		Msg("Returning latest remote configuration")
	return cm[len(cm)-1][1], nil
}
