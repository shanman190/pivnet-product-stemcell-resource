package check_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/shanman190/pivnet-product-stemcell-resource/check"
	"github.com/shanman190/pivnet-product-stemcell-resource/check/checkfakes"
	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"

	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
	"github.com/pivotal-cf/pivnet-resource/v3/versions"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check", func() {
	var (
		fakeLogger       logger.Logger
		fakeFilter       *checkfakes.FakeFilter
		fakePivnetClient *checkfakes.FakePivnetClient
		fakeSorter       *checkfakes.FakeSorter

		checkRequest concourse.CheckRequest
		checkCommand *check.Command

		productSlug	string
		stemcellSlug string
		
		productVersionsWithFingerprints []string
		stemcellVersionsWithFingerprints []string

		releaseTypes    []pivnet.ReleaseType
		releaseTypesErr error

		productReleases         []pivnet.Release
		releasesErr             error
		filteredProductReleases []pivnet.Release
		
		releasesByReleaseTypeErr error
		releasesByVersionErr     error

		allReleaseDependencies	  []pivnet.ReleaseDependency
		allReleaseDependenciesErr error

		stemcellReleases    []pivnet.Release
		stemcellReleasesErr error
		
		tempDir     string
		logFilePath string
	)

	BeforeEach(func() {
		logger := log.New(GinkgoWriter, "", log.LstdFlags)
		fakeLogger = logshim.NewLogShim(logger, logger, true)

		fakeFilter = &checkfakes.FakeFilter{}
		fakePivnetClient = &checkfakes.FakePivnetClient{}
		fakeSorter = &checkfakes.FakeSorter{}

		productSlug = "some product"
		stemcellSlug = "some stemcell"

		releasesByReleaseTypeErr = nil
		releasesByVersionErr = nil
		releaseTypesErr = nil
		releasesErr = nil
		allReleaseDependenciesErr = nil
		stemcellReleasesErr = nil

		releaseTypes = []pivnet.ReleaseType{
			pivnet.ReleaseType("foo release"),
			pivnet.ReleaseType("bar"),
			pivnet.ReleaseType("third release type"),
		}

		productReleases = []pivnet.Release{
			{
				ID:                     1,
				Version:                "1.2.3",
				ReleaseType:            releaseTypes[0],
				SoftwareFilesUpdatedAt: "time1",
			},
			{
				ID:						2,
				Version:				"2.3.4",
				ReleaseType:			releaseTypes[1],
				SoftwareFilesUpdatedAt: "time2",
			},
			{
				ID:						3,
				Version:				"1.2.4",
				ReleaseType: 			releaseTypes[2],
				SoftwareFilesUpdatedAt: "time3",
			},
		}

		allReleaseDependencies = []pivnet.ReleaseDependency{
			{
				Release: pivnet.DependentRelease{
					ID:      21,
					Version: "100.21",
					Product: pivnet.Product{
						ID:   11,
						Slug: "some stemcell",
						Name: "some stemcell name",
					},
				},
			},
			{
				Release: pivnet.DependentRelease{
					ID:      22,
					Version: "210.97",
					Product: pivnet.Product{
						ID:   12,
						Slug: "some stemcell",
						Name: "some stemcell name",
					},
				},
			},
			{
				Release: pivnet.DependentRelease{
					ID:      23,
					Version: "150.64",
					Product: pivnet.Product{
						ID:   13,
						Slug: "some stemcell",
						Name: "some stemcell name",
					},
				},
			},
		}

		stemcellReleases = []pivnet.Release {
			{
				ID: 					11,
				Version: 				"100.21",
				ReleaseType: 			releaseTypes[0],
				SoftwareFilesUpdatedAt: "time1",
			},
			{
				ID:						12,
				Version:				"210.97",
				ReleaseType:			releaseTypes[1],
				SoftwareFilesUpdatedAt: "time2",
			},
			{
				ID:						13,
				Version:				"150.64",
				ReleaseType:			releaseTypes[2],
				SoftwareFilesUpdatedAt: "time3",
			},
		}

		productVersionsWithFingerprints = make([]string, len(productReleases))
		for i, r := range productReleases {
			v, err := versions.CombineVersionAndFingerprint(r.Version, r.SoftwareFilesUpdatedAt)
			Expect(err).NotTo(HaveOccurred())
			productVersionsWithFingerprints[i] = v
		}

		stemcellVersionsWithFingerprints = make([]string, len(stemcellReleases))
		for i, r := range stemcellReleases {
			v, err := versions.CombineVersionAndFingerprint(r.Version, r.SoftwareFilesUpdatedAt)
			Expect(err).NotTo(HaveOccurred())
			stemcellVersionsWithFingerprints[i] = v
		}

		filteredProductReleases = productReleases

		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				APIToken:    "some-api-token",
				ProductSlug: productSlug,
				StemcellSlug: stemcellSlug,
			},
		}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath = filepath.Join(tempDir, "pivnet-resource-check.log1234")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		fakePivnetClient.ReleaseTypesReturns(releaseTypes, releaseTypesErr)
		fakePivnetClient.ReleasesForProductSlugReturns(productReleases, releasesErr)

		fakeFilter.ReleasesByReleaseTypeReturns(filteredProductReleases, releasesByReleaseTypeErr)
		fakeFilter.ReleasesByVersionReturns(filteredProductReleases, releasesByVersionErr)

		binaryVersion := "v0.1.2-unit-tests"

		checkCommand = check.NewCheckCommand(
			fakeLogger,
			binaryVersion,
			fakeFilter,
			fakePivnetClient,
			fakeSorter,
			logFilePath,
		)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when no version data is provided", func() {
		BeforeEach(func() {
			fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[0]}, allReleaseDependenciesErr)
			fakePivnetClient.ReleaseDependenciesReturnsOnCall(1, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)
			fakePivnetClient.ReleaseDependenciesReturnsOnCall(2, []pivnet.ReleaseDependency{allReleaseDependencies[2]}, allReleaseDependenciesErr)
			fakePivnetClient.GetReleaseReturnsOnCall(0, stemcellReleases[0], stemcellReleasesErr)
			fakePivnetClient.GetReleaseReturnsOnCall(1, stemcellReleases[1], stemcellReleasesErr)
			fakePivnetClient.GetReleaseReturnsOnCall(2, stemcellReleases[2], stemcellReleasesErr)
		})

		It("returns the most recent version without error", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			expectedProductVersionWithFingerprint := productVersionsWithFingerprints[0]
			expectedStemcellVersionWithFingerprint := stemcellVersionsWithFingerprints[0]

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(expectedProductVersionWithFingerprint))
			Expect(response[0].StemcellVersion).To(Equal(expectedStemcellVersionWithFingerprint))
		})

		Context("when log files already exist", func() {
			var (
				otherFilePath1 string
				otherFilePath2 string
			)

			BeforeEach(func() {
				otherFilePath1 = filepath.Join(tempDir, "pivnet-resource-check.log1")
				err := ioutil.WriteFile(otherFilePath1, []byte("initial log content"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				otherFilePath2 = filepath.Join(tempDir, "pivnet-resource-check.log2")
				err = ioutil.WriteFile(otherFilePath2, []byte("initial log content"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("removes the other log files", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(otherFilePath1)
				Expect(err).To(HaveOccurred())
				Expect(os.IsNotExist(err)).To(BeTrue())

				_, err = os.Stat(otherFilePath2)
				Expect(err).To(HaveOccurred())
				Expect(os.IsNotExist(err)).To(BeTrue())

				_, err = os.Stat(logFilePath)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("when no releases are returned", func() {
		BeforeEach(func() {
			productReleases = []pivnet.Release{}
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("cannot find specified product release"))
		})
	})

	Context("when there is an error getting release types", func() {
		BeforeEach(func() {
			releaseTypesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("some error"))
		})
	})

	Context("when there is an error getting releases", func() {
		BeforeEach(func() {
			releasesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("some error"))
		})
	})

	Describe("when a version is provided", func() {
		Context("when the version is the latest", func() {
			BeforeEach(func() {
				productVersionWithFingerprint := productVersionsWithFingerprints[0]
				stemcellVersionWithFingerprint := stemcellVersionsWithFingerprints[0]

				checkRequest.Version = concourse.Version{
					ProductVersion: productVersionWithFingerprint,
					StemcellVersion: stemcellVersionWithFingerprint,
				}

				fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[0]}, allReleaseDependenciesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(0, stemcellReleases[0], stemcellReleasesErr)
			})

			It("returns the most recent version", func() {
				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				productVersionWithFingerprintA := productVersionsWithFingerprints[0]
				stemcellVersionWithFingerprintA := stemcellVersionsWithFingerprints[0]

				Expect(response).To(HaveLen(1))
				Expect(response[0].ProductVersion).To(Equal(productVersionWithFingerprintA))
				Expect(response[0].StemcellVersion).To(Equal(stemcellVersionWithFingerprintA))
			})
		})

		Context("when the version is not the latest", func() {
			BeforeEach(func() {
				productVersionWithFingerprint := productVersionsWithFingerprints[2] // 1.2.4#time3
				stemcellVersionWithFingerprint := stemcellVersionsWithFingerprints[2] // 150.64#time3

				checkRequest.Version = concourse.Version{
					ProductVersion: productVersionWithFingerprint,
					StemcellVersion: stemcellVersionWithFingerprint,
				}

				fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[0]}, allReleaseDependenciesErr)
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(1, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(2, []pivnet.ReleaseDependency{allReleaseDependencies[2]}, allReleaseDependenciesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(0, stemcellReleases[0], stemcellReleasesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(1, stemcellReleases[1], stemcellReleasesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(2, stemcellReleases[2], stemcellReleasesErr)
			})

			It("returns the most recent versions, including the version specified", func() {
				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				productVersionWithFingerprintA := productVersionsWithFingerprints[0] // 1.2.3#time1
				stemcellVersionWithFingerprintA := stemcellVersionsWithFingerprints[0] // 100.21#time1
				productVersionWithFingerprintB := productVersionsWithFingerprints[1] // 2.3.4#time2
				stemcellVersionWithFingerprintB := stemcellVersionsWithFingerprints[1] // 210.97#time2
				productVersionWithFingerprintC := productVersionsWithFingerprints[2] // 1.2.4#time3
				stemcellVersionWithFingerprintC := stemcellVersionsWithFingerprints[2] // 150.64#time3

				Expect(response).To(HaveLen(3))
				Expect(response[0].ProductVersion).To(Equal(productVersionWithFingerprintC))
				Expect(response[0].StemcellVersion).To(Equal(stemcellVersionWithFingerprintC))
				Expect(response[1].ProductVersion).To(Equal(productVersionWithFingerprintB))
				Expect(response[1].StemcellVersion).To(Equal(stemcellVersionWithFingerprintB))
				Expect(response[2].ProductVersion).To(Equal(productVersionWithFingerprintA))
				Expect(response[2].StemcellVersion).To(Equal(stemcellVersionWithFingerprintA))
			})
		})
	})

	Context("when the release type is specified", func() {
		BeforeEach(func() {
			checkRequest.Source.ReleaseType = string(releaseTypes[1])

			filteredProductReleases = []pivnet.Release{productReleases[1]}

			fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)
			fakePivnetClient.GetReleaseReturnsOnCall(0, stemcellReleases[1], stemcellReleasesErr)
		})

		It("returns the most recent version with that release type", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			productVersionWithFingerprintC := productVersionsWithFingerprints[1]
			stemcellVersionWithFingerprintC := stemcellVersionsWithFingerprints[1]

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(productVersionWithFingerprintC))
			Expect(response[0].StemcellVersion).To(Equal(stemcellVersionWithFingerprintC))
		})

		Context("when the release type is invalid", func() {
			BeforeEach(func() {
				checkRequest.Source.ReleaseType = "not a valid release type"
			})

			It("returns an error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*release type.*one of"))
				Expect(err.Error()).To(ContainSubstring(string(releaseTypes[0])))
				Expect(err.Error()).To(ContainSubstring(string(releaseTypes[1])))
				Expect(err.Error()).To(ContainSubstring(string(releaseTypes[2])))
			})
		})

		Context("when filtering returns an error", func() {
			BeforeEach(func() {
				releasesByReleaseTypeErr = fmt.Errorf("some release type error")
			})

			It("returns the error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(releasesByReleaseTypeErr))
			})
		})
	})

	Context("when the product version is specified", func() {
		BeforeEach(func() {
			checkRequest.Source.ReleaseType = string(releaseTypes[1])

			filteredProductReleases = []pivnet.Release{productReleases[1]}

			fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)
			fakePivnetClient.GetReleaseReturnsOnCall(0, stemcellReleases[1], stemcellReleasesErr)
		})

		BeforeEach(func() {
			checkRequest.Source.ProductVersion = "C"
		})

		It("returns the newest release with that version without error", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			productVersionWithFingerprintC := productVersionsWithFingerprints[1]

			Expect(response).To(HaveLen(1))
			Expect(response[0].ProductVersion).To(Equal(productVersionWithFingerprintC))
		})

		Context("when filtering returns an error", func() {
			BeforeEach(func() {
				releasesByVersionErr = fmt.Errorf("some version error")
			})

			It("returns the error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(releasesByVersionErr))
			})
		})
	})

	Context("when sorting by semver", func() {
		var (
			semverOrderedProductReleases []pivnet.Release
		)

		BeforeEach(func() {
			checkRequest.Source.SortBy = concourse.SortBySemver

			semverOrderedProductReleases = []pivnet.Release{
				productReleases[1], // 2.3.4
				productReleases[2], // 1.2.4
				productReleases[0], // 1.2.3
			}

			checkRequest.Version = concourse.Version{
				ProductVersion: productVersionsWithFingerprints[0], // 1.2.3#time1
				StemcellVersion: stemcellVersionsWithFingerprints[0], // 100.21#time1
			}
		})

		Context("when no errors are raised", func() {
			BeforeEach(func() {
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(1, []pivnet.ReleaseDependency{allReleaseDependencies[2], allReleaseDependencies[0]}, allReleaseDependenciesErr)
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(2, []pivnet.ReleaseDependency{allReleaseDependencies[0]}, allReleaseDependenciesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(0, stemcellReleases[1], stemcellReleasesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(1, stemcellReleases[2], stemcellReleasesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(2, stemcellReleases[0], stemcellReleasesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(3, stemcellReleases[0], stemcellReleasesErr)

				fakeSorter.SortBySemverReturnsOnCall(0, semverOrderedProductReleases, nil)
				fakeSorter.SortBySemverReturnsOnCall(1, []pivnet.Release{stemcellReleases[1]}, nil)
				fakeSorter.SortBySemverReturnsOnCall(2, []pivnet.Release{stemcellReleases[2], stemcellReleases[0]}, nil)
				fakeSorter.SortBySemverReturnsOnCall(3, []pivnet.Release{stemcellReleases[0]}, nil)
			})

			It("returns in ascending semver order", func() {
				Expect(fakeSorter.SortBySemverCallCount()).To(Equal(0))

				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(4))
				Expect(response[0].ProductVersion).To(Equal(productVersionsWithFingerprints[0]))
				Expect(response[0].StemcellVersion).To(Equal(stemcellVersionsWithFingerprints[0]))
				Expect(response[1].ProductVersion).To(Equal(productVersionsWithFingerprints[2]))
				Expect(response[1].StemcellVersion).To(Equal(stemcellVersionsWithFingerprints[0]))
				Expect(response[2].ProductVersion).To(Equal(productVersionsWithFingerprints[2]))
				Expect(response[2].StemcellVersion).To(Equal(stemcellVersionsWithFingerprints[2]))
				Expect(response[3].ProductVersion).To(Equal(productVersionsWithFingerprints[1]))
				Expect(response[3].StemcellVersion).To(Equal(stemcellVersionsWithFingerprints[1]))

				Expect(fakeSorter.SortBySemverCallCount()).To(Equal(4))
			})
		})

		Context("when sorting products by semver returns an error", func() {
			var (
				semverErr error
			)

			BeforeEach(func() {
				semverErr = errors.New("semver error")

				fakeSorter.SortBySemverReturns(nil, semverErr)
			})

			It("returns error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(semverErr))
			})
		})

		Context("when sorting stemcells by semver returns an error", func() {
			var (
				semverErr error
			)

			BeforeEach(func() {
				semverErr = errors.New("semver error")

				fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)

				fakeSorter.SortBySemverReturnsOnCall(0, semverOrderedProductReleases, nil)
				fakeSorter.SortBySemverReturnsOnCall(1, nil, semverErr)
			})

			It("returns error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(semverErr))
			})
		})
	})

	Context("when sorting by last_updated", func() {
		var (
			updateSortedProductReleases []pivnet.Release
		)

		BeforeEach(func() {
			checkRequest.Source.SortBy = concourse.SortByLastUpdated

			updateSortedProductReleases = []pivnet.Release{
				productReleases[2], // 1.2.4#time3
				productReleases[1], // 2.3.4#time2
				productReleases[0], // 1.2.3#time1
			}

			checkRequest.Version = concourse.Version{
				ProductVersion: productVersionsWithFingerprints[0], // 1.2.3#time1
				StemcellVersion: stemcellVersionsWithFingerprints[0], // 100.24#time1
			}
		})

		Context("when no errors are raised", func() {
			BeforeEach(func() {
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[2]}, allReleaseDependenciesErr)
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(1, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)
				fakePivnetClient.ReleaseDependenciesReturnsOnCall(2, []pivnet.ReleaseDependency{allReleaseDependencies[0]}, allReleaseDependenciesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(0, stemcellReleases[2], stemcellReleasesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(1, stemcellReleases[1], stemcellReleasesErr)
				fakePivnetClient.GetReleaseReturnsOnCall(2, stemcellReleases[0], stemcellReleasesErr)

				fakeSorter.SortByLastUpdatedReturnsOnCall(0, updateSortedProductReleases, nil)
				fakeSorter.SortByLastUpdatedReturnsOnCall(1, []pivnet.Release{stemcellReleases[2]}, nil)
				fakeSorter.SortByLastUpdatedReturnsOnCall(2, []pivnet.Release{stemcellReleases[1]}, nil)
				fakeSorter.SortByLastUpdatedReturnsOnCall(3, []pivnet.Release{stemcellReleases[0]}, nil)
			})

			It("returns invokes sort by last_updated on sorter", func() {
				Expect(fakeSorter.SortByLastUpdatedCallCount()).To(Equal(0))

				response, err := checkCommand.Run(checkRequest)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(HaveLen(3))
				Expect(response[0].ProductVersion).To(Equal(productVersionsWithFingerprints[0]))
				Expect(response[0].StemcellVersion).To(Equal(stemcellVersionsWithFingerprints[0]))
				Expect(response[1].ProductVersion).To(Equal(productVersionsWithFingerprints[1]))
				Expect(response[1].StemcellVersion).To(Equal(stemcellVersionsWithFingerprints[1]))
				Expect(response[2].ProductVersion).To(Equal(productVersionsWithFingerprints[2]))
				Expect(response[2].StemcellVersion).To(Equal(stemcellVersionsWithFingerprints[2]))

				Expect(fakeSorter.SortByLastUpdatedCallCount()).To(Equal(4))
			})
		})

		Context("when sorting products by last_updated returns an error", func() {
			var (
				semverErr error
			)

			BeforeEach(func() {
				semverErr = errors.New("semver error")

				fakeSorter.SortByLastUpdatedReturns(nil, semverErr)
			})

			It("returns error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(semverErr))
			})
		})

		Context("when sorting stemcells by last_updated returns an error", func() {
			var (
				semverErr error
			)

			BeforeEach(func() {
				semverErr = errors.New("semver error")

				fakePivnetClient.ReleaseDependenciesReturnsOnCall(0, []pivnet.ReleaseDependency{allReleaseDependencies[1]}, allReleaseDependenciesErr)

				fakeSorter.SortByLastUpdatedReturnsOnCall(0, updateSortedProductReleases, nil)
				fakeSorter.SortByLastUpdatedReturnsOnCall(1, nil, semverErr)
			})

			It("returns error", func() {
				_, err := checkCommand.Run(checkRequest)
				Expect(err).To(HaveOccurred())

				Expect(err).To(Equal(semverErr))
			})
		})
	})
})