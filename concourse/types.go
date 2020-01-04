package concourse

// SortBy : type alias for better readability
type SortBy string

const (
	// SortByNone : Use sorting as-is from PivNet API
	SortByNone   	  SortBy = "none"
	// SortBySemver : Sort the responses by Semantic Versioning rules
	SortBySemver 	  SortBy = "semver"
	// SortByLastUpdated : Sort the responses by their last updated time
	SortByLastUpdated SortBy = "last_updated"
)

// Source : source structure for information provided from Concourse
type Source struct {
	APIToken          string `json:"api_token"`
	ProductSlug       string `json:"product_slug"`
	ProductVersion    string `json:"product_version"`
	StemcellSlug	  string `json:"stemcell_slug"`
	Endpoint          string `json:"endpoint"`
	ReleaseType       string `json:"release_type"`
	SortBy            SortBy `json:"sort_by"`
	SkipSSLValidation bool   `json:"skip_ssl_verification"`
	CopyMetadata      bool   `json:"copy_metadata"`
	Verbose           bool   `json:"verbose"`
}

// CheckRequest : request body for the check.Command
type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

// Version : version structure for information provided in Concourse
type Version struct {
	ProductVersion string `json:"product_version"`
	StemcellVersion string `json:"stemcell_version"`
}

// CheckResponse : response body for the check.Command
type CheckResponse []Version

// InRequest : request body for the in.Command
type InRequest struct {
	Source  Source   `json:"source"`
	Version Version  `json:"version"`
	Params  InParams `json:"params"`
}

// InParams : parameter structure for information provided from Concourse on get usages
type InParams struct {
	Globs  []string `json:"globs"`
	Unpack bool     `json:"unpack"`
}

// InResponse : response body for the in.Command
type InResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata,omitempty"`
}

// Metadata : metadata structure for information provided in Concourse
type Metadata struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}