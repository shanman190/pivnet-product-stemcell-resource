package main

import (
	"encoding/json"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
	"github.com/pivotal-cf/pivnet-resource/v3/filter"
	"github.com/pivotal-cf/pivnet-resource/v3/gp"
	"github.com/pivotal-cf/pivnet-resource/v3/semver"
	"github.com/pivotal-cf/pivnet-resource/v3/sorter"
	"github.com/pivotal-cf/pivnet-resource/v3/useragent"
	"github.com/robdimsdale/sanitizer"
	"io/ioutil"
	"log"
	"os"
	"github.com/shanman190/pivnet-product-stemcell-resource/check"
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

	var input concourse.CheckRequest

	logFile, err := ioutil.TempFile("", "pivnet-check.log")
	if err != nil {
		log.Printf("could not create log file")
	}

	logger := log.New(logFile, "", log.LstdFlags)

	logger.Printf("PivNet Product Stemcell Resource version: %s", version)

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	logger.SetOutput(sanitizer.NewSanitizer(sanitized, logFile))

	verbose := false
	ls := logshim.NewLogShim(logger, logger, verbose)

	err = validator.NewCheckValidator(input).Validate()
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	var endpoint string
	if input.Source.Endpoint != "" {
		endpoint = input.Source.Endpoint
	} else {
		endpoint = pivnet.DefaultHost
	}

	apiToken := input.Source.APIToken
	token := pivnet.NewAccessTokenOrLegacyToken(apiToken, endpoint, input.Source.SkipSSLValidation, "Pivnet Product Stemcell Resource")

	client := newPivnetClientWithToken(
		token,
		endpoint,
		input.Source.SkipSSLValidation,
		useragent.UserAgent(version, "check", input.Source.ProductSlug),
		ls,
	)

	f := filter.NewFilter(ls)

	semverConverter := semver.NewSemverConverter(ls)
	s := sorter.NewSorter(ls, semverConverter)

	response, err := check.NewCheckCommand(
		ls,
		version,
		f,
		client,
		s,
		logFile.Name(),
	).Run(input)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		log.Fatalf("Exiting with error: %s", err)
	}
}

func newPivnetClientWithToken(token pivnet.AccessTokenOrLegacyToken, host string, skipSSLValidation bool, userAgent string, logger logger.Logger) *gp.Client {
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