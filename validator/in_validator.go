package validator

import (
	"fmt"

	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
)

type InValidator struct {
	input concourse.InRequest
}

func NewInValidator(input concourse.InRequest) *InValidator {
	return &InValidator{
		input: input,
	}
}

func (v InValidator) Validate() error {
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}

	if v.input.Source.StemcellSlug == "" {
		return fmt.Errorf("%s must be provided", "stemcell_slug")
	}

	if v.input.Version.ProductVersion == "" {
		return fmt.Errorf("%s must be provided", "product_version")
	}

	if v.input.Version.StemcellVersion == "" {
		return fmt.Errorf("%s must be provided", "stemcell_version")
	}

	return nil
}
