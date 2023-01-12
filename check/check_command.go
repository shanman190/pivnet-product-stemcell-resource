package check

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	
	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
	"github.com/shanman190/pivnet-product-stemcell-resource/versions"
)

//go:generate counterfeiter --fake-name FakeFilter . filter
type filter interface {
	ReleasesByReleaseType(releases []pivnet.Release, releaseType pivnet.ReleaseType) ([]pivnet.Release, error)
	ReleasesByVersion(releases []pivnet.Release, version string) ([]pivnet.Release, error)
}

//go:generate counterfeiter --fake-name FakeSorter . sorter
type sorter interface {
	SortBySemver([]pivnet.Release) ([]pivnet.Release, error)
	SortByLastUpdated([]pivnet.Release) ([]pivnet.Release, error)
}

//go:generate counterfeiter --fake-name FakePivnetClient . pivnetClient
type pivnetClient interface {
	ReleaseTypes() ([]pivnet.ReleaseType, error)
	ReleasesForProductSlug(string) ([]pivnet.Release, error)
	ReleaseDependencies(string, int) ([]pivnet.ReleaseDependency, error)
	GetRelease(string, string) (pivnet.Release, error)
}

// Command : Concourse Check command to search for new versions.
type Command struct {
	logger        logger.Logger
	binaryVersion string
	filter        filter
	pivnetClient  pivnetClient
	sort		  sorter
	logFilePath   string
}

// NewCheckCommand : Creates an instance of the check Command
func NewCheckCommand(
	logger logger.Logger,
	binaryVersion string,
	filter filter,
	pivnetClient pivnetClient,
	sort sorter,
	logFilePath string,
) *Command {
	return &Command{
		logger:        logger,
		binaryVersion: binaryVersion,
		filter:        filter,
		pivnetClient:  pivnetClient,
		sort: 		   sort,
		logFilePath:   logFilePath,
	}
}

// Run : Execute the check command
func (c *Command) Run(input concourse.CheckRequest) (concourse.CheckResponse, error) {
	c.logger.Info("Received input, starting Check CMD run")

	err := c.removeExistingLogFiles()
	if err != nil {
		return nil, err
	}

	releaseType := input.Source.ReleaseType

	err = c.validateReleaseType(releaseType)
	if err != nil {
		return nil, err
	}

	productSlug := input.Source.ProductSlug
	stemcellSlug := input.Source.StemcellSlug

	c.logger.Info("Getting all product releases")
	productReleases, err := c.pivnetClient.ReleasesForProductSlug(productSlug)
	if err != nil {
		return nil, err
	}

	if releaseType != "" {
		c.logger.Info(fmt.Sprintf("Filtering all product releases by release type: '%s'", releaseType))
		productReleases, err = c.filter.ReleasesByReleaseType(
			productReleases,
			pivnet.ReleaseType(releaseType),
		)
		if err != nil {
			return nil, err
		}
	}

	productVersion := input.Source.ProductVersion
	if productVersion != "" {
		c.logger.Info(fmt.Sprintf("Filtering all product releases by product version: '%s'", productVersion))
		productReleases, err = c.filter.ReleasesByVersion(productReleases, productVersion)
		if err != nil {
			return nil, err
		}
	}

	if input.Source.SortBy == concourse.SortBySemver {
		c.logger.Info("Sorting all product releases by semver")
		productReleases, err = c.sort.SortBySemver(productReleases)
		if err != nil {
			return nil, err
		}
	} else if input.Source.SortBy == concourse.SortByLastUpdated {
		c.logger.Info("Sorting all product releases by release date")
		productReleases, err = c.sort.SortByLastUpdated(productReleases)
		if err != nil {
			return nil, err
		}
	}

	if len(productReleases) == 0 {
		return concourse.CheckResponse{}, fmt.Errorf("cannot find specified product release")
	}

	c.logger.Info("Gathering new product versions")

	lastSeenProductVersion, _, _ := versions.SplitIntoVersionAndFingerprint(input.Version.ProductVersion)
	newProductReleases, err := versions.SinceRelease(productReleases, lastSeenProductVersion)
	if err != nil {
		// Untested because versions.Since cannot be forced to return an error.
		return nil, err
	}

	c.logger.Info("Gathering new stemcell versions")

	productsToStemcells := make(map[string][]string)
	stemcellCache := make(map[string]pivnet.Release)
	for _, productRelease := range newProductReleases {
		c.logger.Info(fmt.Sprintf("Getting release dependencies for '%s/%s'", productSlug, productRelease.Version))
		releaseDependencies, err := c.pivnetClient.ReleaseDependencies(productSlug, productRelease.ID)
		if err != nil {
			return nil, err
		}

		var stemcellVersions []string
		for _, productReleaseDependency := range releaseDependencies {
			if strings.Contains(productReleaseDependency.Release.Product.Slug, stemcellSlug) {
				stemcellVersions = append(stemcellVersions, productReleaseDependency.Release.Version)
			}
		}

		lastSeenStemcellVersion, _, _ := versions.SplitIntoVersionAndFingerprint(input.Version.StemcellVersion)
		newStemcellVersions, err := versions.Since(stemcellVersions, lastSeenStemcellVersion)
		if err != nil {
			// Untested because versions.Since cannot be forced to return an error.
			return nil, err
		}

		var stemcells []pivnet.Release
		for _, stemcellVersion := range newStemcellVersions {
			c.logger.Info(fmt.Sprintf("Getting release details for '%s/%s'", stemcellSlug, stemcellVersion))
			stemcellRelease, ok := stemcellCache[stemcellVersion]
			if !ok {
				stemcellRelease, err = c.pivnetClient.GetRelease(stemcellSlug, stemcellVersion)
				if err != nil {
					return nil, err
				}

				stemcellCache[stemcellVersion] = stemcellRelease
			}

			stemcells = append(stemcells, stemcellRelease)
		}

		if input.Source.SortBy == concourse.SortBySemver {
			c.logger.Info("Sorting all stemcell releases by semver")
			stemcells, err = c.sort.SortBySemver(stemcells)
			if err != nil {
				return nil, err
			}
		} else if input.Source.SortBy == concourse.SortByLastUpdated {
			c.logger.Info("Sorting all stemcell releases by release date")
			stemcells, err = c.sort.SortByLastUpdated(stemcells)
			if err != nil {
				return nil, err
			}
		}

		fingerprintedStemcellVersions, err := releaseVersions(stemcells)
		if err != nil {
			// Untested because versions.CombineVersionAndFingerprint cannot be forced to return an error.
			return concourse.CheckResponse{}, err
		}

		if len(fingerprintedStemcellVersions) == 0 {
			return concourse.CheckResponse{}, fmt.Errorf("cannot find specified stemcell release")
		}

		fingerprintedProductVersion, err := versions.CombineVersionAndFingerprint(productRelease.Version, productRelease.SoftwareFilesUpdatedAt)
		if err != nil {
			return nil, err
		}
		productsToStemcells[fingerprintedProductVersion] = fingerprintedStemcellVersions
	}

	c.logger.Info(fmt.Sprintf("New versions: %v", productsToStemcells))

	var out concourse.CheckResponse
	productVersions, err := releaseVersions(productReleases)
	if err != nil {
		return nil, err
	}
	reversedProductVersions, err := versions.Reverse(productVersions)
	if err != nil {
		// Untested because versions.Reverse cannot be forced to return an error.
		return nil, err
	}

	for _, pv := range reversedProductVersions {
		reversedStemcellVersions, err := versions.Reverse(productsToStemcells[pv])
		if err != nil {
			return nil, err
		}

		for _, sv := range reversedStemcellVersions {
			out = append(out, concourse.Version{ProductVersion: pv, StemcellVersion: sv})
		}
	}

	c.logger.Info("Finishing check and returning output")

	return out, nil
}

func (c *Command) removeExistingLogFiles() error {
	logDir := filepath.Dir(c.logFilePath)
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "*.log*"))
	if err != nil {
		// This is untested because the only error returned by filepath.Glob is a
		// malformed glob, and this glob is hard-coded to be correct.
		return err
	}

	c.logger.Info(fmt.Sprintf("Located logfiles: %v", existingLogFiles))

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(c.logFilePath) {
			c.logger.Info(fmt.Sprintf("Removing existing log file: %s", f))
			err := os.Remove(f)
			if err != nil {
				// This is untested because it is too hard to force os.Remove to return
				// an error.
				return err
			}
		}
	}

	return nil
}

func (c *Command) validateReleaseType(releaseType string) error {
	c.logger.Info(fmt.Sprintf("Validating release type: '%s'", releaseType))
	releaseTypes, err := c.pivnetClient.ReleaseTypes()
	if err != nil {
		return err
	}

	releaseTypesAsStrings := make([]string, len(releaseTypes))
	for i, r := range releaseTypes {
		releaseTypesAsStrings[i] = string(r)
	}

	if releaseType != "" && !containsString(releaseTypesAsStrings, releaseType) {
		releaseTypesPrintable := fmt.Sprintf("['%s']", strings.Join(releaseTypesAsStrings, "', '"))
		return fmt.Errorf(
			"provided release type: '%s' must be one of: %s",
			releaseType,
			releaseTypesPrintable,
		)
	}

	return nil
}

func containsString(strings []string, str string) bool {
	for _, s := range strings {
		if str == s {
			return true
		}
	}
	return false
}

func releaseVersions(releases []pivnet.Release) ([]string, error) {
	releaseVersions := make([]string, len(releases))

	var err error
	for i, r := range releases {
		releaseVersions[i], err = versions.CombineVersionAndFingerprint(r.Version, r.SoftwareFilesUpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return releaseVersions, nil
}