package main

import (
	"encoding/json"
	"github.com/shanman190/pivnet-product-stemcell-resource/versions"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/pivotal-cf/go-pivnet/v3"
	"github.com/pivotal-cf/go-pivnet/v3/logger"
	"github.com/pivotal-cf/go-pivnet/v3/logshim"
	"github.com/pivotal-cf/go-pivnet/v3/md5sum"
	"github.com/pivotal-cf/go-pivnet/v3/sha256sum"
	pivnetconcourse "github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/downloader"
	"github.com/pivotal-cf/pivnet-resource/filter"
	"github.com/pivotal-cf/pivnet-resource/gp"
	"github.com/pivotal-cf/pivnet-resource/in"
	"github.com/pivotal-cf/pivnet-resource/in/filesystem"
	"github.com/pivotal-cf/pivnet-resource/ui"
	"github.com/pivotal-cf/pivnet-resource/useragent"
	"github.com/robdimsdale/sanitizer"

	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
	"github.com/shanman190/pivnet-product-stemcell-resource/validator"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string
)

func main() {
	if version == "" {
		version = "dev"
	}

	color.NoColor = false

	logWriter := os.Stderr
	uiPrinter := ui.NewUIPrinter(logWriter)

	logger := log.New(logWriter, "", log.LstdFlags)

	logger.Printf("PivNet Product Stemcell Resource version: %s", version)

	if len(os.Args) < 2 {
		uiPrinter.PrintErrorlnf(
			"not enough args - usage: %s <sources directory>",
			os.Args[0],
		)
		os.Exit(1)
	}

	downloadDir := os.Args[1]

	var input concourse.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	logger.SetOutput(sanitizer.NewSanitizer(sanitized, logWriter))

	verbose := input.Source.Verbose
	ls := logshim.NewLogShim(logger, logger, verbose)

	ls.Debug("Verbose output enabled")
	logger.Printf("Creating download directory: %s", downloadDir)

	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	err = validator.NewInValidator(input).Validate()
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.DefaultHost
	}

	apiToken := input.Source.APIToken
	token := pivnet.NewAccessTokenOrLegacyToken(apiToken, endpoint, input.Source.SkipSSLValidation, "Pivnet Resource")

	if len(apiToken) < 20 {
		uiPrinter.PrintDeprecationln("The use of static Pivnet API tokens is deprecated and will be removed. Please see https://network.pivotal.io/docs/api#how-to-authenticate for details.")
	}

	client := NewPivnetClientWithToken(
		token,
		endpoint,
		input.Source.SkipSSLValidation,
		useragent.UserAgent(version, "get", input.Source.ProductSlug),
		ls,
	)

	d := downloader.NewDownloader(client, downloadDir, ls, logWriter)

	fs := sha256sum.NewFileSummer()
	md5fs := md5sum.NewFileSummer()

	f := filter.NewFilter(ls)

	fileWriter := filesystem.NewFileWriter(downloadDir, ls)
	archive := &in.Archive{}

	pivnetInput := convertToPivnetInput(input)

	pivnetResponse, err := in.NewInCommand(
		ls,
		client,
		f,
		d,
		fs,
		md5fs,
		fileWriter,
		archive,
	).Run(pivnetInput)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}

	response := convertFromPivnetResponse(input.Version.ProductVersion, pivnetResponse)

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		uiPrinter.PrintErrorln(err)
		os.Exit(1)
	}
}

func NewPivnetClientWithToken(token pivnet.AccessTokenService, host string, skipSSLValidation bool, userAgent string, logger logger.Logger) *gp.Client {
	clientConfig := pivnet.ClientConfig{
		Host:              host,
		UserAgent:         userAgent,
		SkipSSLValidation: skipSSLValidation,
	}

	return gp.NewClient(
		token,
		clientConfig,
		logger,
	)
}

func convertToPivnetInput(input concourse.InRequest) pivnetconcourse.InRequest {
	stemcellVersion, _, _ := versions.SplitIntoVersionAndFingerprint(input.Version.StemcellVersion)

	var pivnetSortyBy pivnetconcourse.SortBy
	if input.Source.SortBy == concourse.SortByNone {
		pivnetSortyBy = pivnetconcourse.SortByNone
	} else if input.Source.SortBy == concourse.SortBySemver {
		pivnetSortyBy = pivnetconcourse.SortBySemver
	} else if input.Source.SortBy == concourse.SortByLastUpdated {
		pivnetSortyBy = pivnetconcourse.SortByLastUpdated
	}

	return pivnetconcourse.InRequest{
		Source:  pivnetconcourse.Source{
			APIToken: 		   input.Source.APIToken,
			ProductSlug: 	   input.Source.StemcellSlug,
			ProductVersion:    stemcellVersion,
			Endpoint:		   input.Source.Endpoint,
			ReleaseType:	   input.Source.ReleaseType,
			SortBy:			   pivnetSortyBy,
			SkipSSLValidation: input.Source.SkipSSLValidation,
			CopyMetadata:	   input.Source.CopyMetadata,
			Verbose: 		   input.Source.Verbose,
		},
		Version: pivnetconcourse.Version{
			ProductVersion: input.Version.StemcellVersion,
		},
		Params:  pivnetconcourse.InParams{
			Globs:  input.Params.Globs,
			Unpack: input.Params.Unpack,
		},
	}
}

func convertFromPivnetResponse(productVersion string, response pivnetconcourse.InResponse) concourse.InResponse {
	var metadata []concourse.Metadata
	for _, pivnetMetadata := range response.Metadata {
		metadata = append(metadata, concourse.Metadata{
			Name:  pivnetMetadata.Name,
			Value: pivnetMetadata.Value,
		})
	}
	return concourse.InResponse{
		Version:  concourse.Version{
			ProductVersion: productVersion,
			StemcellVersion: response.Version.ProductVersion,
		},
		Metadata: metadata,
	}
}