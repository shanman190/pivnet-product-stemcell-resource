package validator

import (
	"fmt"

	"github.com/shanman190/pivnet-product-stemcell-resource/concourse"
)

type CheckValidator struct {
	input concourse.CheckRequest
}

func NewCheckValidator(input concourse.CheckRequest) *CheckValidator {
	return &CheckValidator{
		input: input,
	}
}

func (v CheckValidator) Validate() error {
	if v.input.Source.APIToken == "" {
		return fmt.Errorf("%s must be provided", "api_token")
	}

	if v.input.Source.ProductSlug == "" {
		return fmt.Errorf("%s must be provided", "product_slug")
	}

	if v.input.Source.StemcellSlug == "" {
		return fmt.Errorf("%s must be provided", "stemcell_slug")
	}
	return nil
}
