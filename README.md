# PivNet Product Stemcell Resource

[![Go Version](https://img.shields.io/github/go-mod/go-version/shanman190/pivnet-product-stemcell-resource)](https://github.com/shanman190/pivnet-product-stemcell-resource)
[![Actions Status](https://img.shields.io/github/workflow/status/shanman190/pivnet-product-stemcell-resource/CI)](https://github.com/shanman190/pivnet-product-stemcell-resource/actions)
[![Code Coverage](https://img.shields.io/codecov/c/github/shanman190/pivnet-product-stemcell-resource)](https://codecov.io/github/shanman190/pivnet-product-stemcell-resource?branch=master)

Inspired by the [PivNet Resource](https://github.com/pivotal-cf/pivnet-resource).

This resource is specifically designed to discover the stemcell dependency of a tracked product. 

## Installing

The recommended method to use this resource is with
[resource_types](https://concourse-ci.org/resource-types.html) in the
pipeline config as follows:

```yaml
---
resource_types:
- name: pivnet-product-stemcell
  type: docker-image
  source:
    repository: shanman190/pivnet-product-stemcell-resource
    tag: latest-final
```

Using `tag: latest-final` will automatically pull the latest final release,
which can be found on the
[releases page](https://github.com/shanman190/pivnet-product-stemcell-resource/releases).

**To avoid automatically upgrading, use a fixed tag instead e.g. `tag: v0.1.0`**

Releases are semantically versioned; these correspond to the git tags in this
repository.

## Source configuration

```yaml
resources:
- name: p-mysql
  type: pivnet
  source:
    api_token: {{api-token}}
    product_slug: p-mysql
    stemcell_slug: stemcells-ubuntu-xenial
```

* `api_token`: *Required string.*

  Token from your Pivotal Network profile. Accepts either your Legacy API Token or UAA Refresh Token.

* `product_slug`: *Required string.*

  Name of product on Pivotal Network.
  
* `stemcell_slug`: *Required string.*

  Name of stemcell on Pivotal Network.

* `release_type`: *Optional boolean.*

  If `true`, lock to a specific Pivotal Network [release type](https://network.pivotal.io/docs/api#releases).

* `copy_metadata`: *Optional boolean.*

  If `true`, copy specified metadata from the latest All Users release within the minor releases. Defaults to `false`.
  
  The following metadata is copied:

  * Release Notes URL
  * End of General Support
  * End of Technical Guidance
  * End of Availability
  * EULA
  * License Exception
  * ECCN
  * Controlled
  * Dependency Specifiers
  * Upgrade Path Specifiers

* `endpoint`: *Optional string.*

  Endpoint to use for communicating with Pivotal Network.

  Defaults to `https://network.pivotal.io`.

* `product_version`: *Optional string.*

  Regular expression to match against product versions, e.g. `1\.2\..*`.

  Empty values match all product versions.

* `sort_by`: *Optional string.*

  Order to use for sorting releases. One of the following:

  - `none`: the order they come back from Pivotal Network.
  - `semver`: by semantic version, in descending order from the highest-valued version.
  - `last_updated`: by last updated at time, in descending order from the most recently updated version. Please note that if an earlier release is updated then the Pivnet Resource 'check' step will return it again. 

## Example pipeline configuration

See [example pipeline configurations](https://github.com/shanman190/pivnet-product-stemcell-resource/blob/master/examples).

## Behavior

### `check`: check for new stemcell versions for the tracked product versions on Pivotal Network

Discovers all stemcell versions of the provided product.
Returned versions are optionally filtered and ordered by the `source` configuration.

### `in`: download the stemcell for the tracked product from Pivotal Network

Downloads the stemcell for the tracked product from Pivotal Network. You will be required to accept a
EULA for any stemcell you're downloading for the first time, as well as if the terms and
conditions associated with the product change.

The metadata for the stemcell is written to both `metadata.json` and
`metadata.yaml` in the working directory (typically `/tmp/build/get`).
Use this to programmatically determine metadata of the release.

See [metadata](https://github.com/shanman190/pivnet-product-stemcell-resource/blob/master/metadata)
for more details on the structure of the metadata file.

#### Parameters

* `globs`: *Optional array.*

  Array of globs matching files to download.

  If multiple files are matched, they are all downloaded.
  
  - The globs match on the actual *file names*, not the display names in Pivotal
  Network. This is to provide a more consistent experience between uploading and
  downloading files.
  
  - If the globs fail to match any files the release download fails
  with error.
  
  - If one or more globs fails to match any files, only the matched files will be downloaded.
  
  - If `globs` is not provided (or is nil), **all files will be downloaded**.
  
  - Setting `globs` to the empty array (i.e. `globs: []`) will not attempt to
  download any files.
  
  - Files are downloaded to the working directory (e.g. `/tmp/build/get`) and the
  file names will be the same as they are on Pivotal Network - e.g. a file with
  name `some-file.txt` will be downloaded to `/tmp/build/get/some-file.txt`.

* `unpack`: *Optional boolean.*

  If `true`, unpack the downloaded file.
  
  - This can be used to use a root filesystem that is packaged as an archive file on
  network.pivotal.io as the image to run a given Concourse task.

More generally, the `unpack` parameter can be used with `get` to pass an image to a task definition,
as in the below example.

```yaml
resource:
- name: image
  type: pivnet
  source:
    api_token: {{pivnet_token}}
    product_slug: {{image-slug}}
    product_version: 0\.0\..*

jobs:
- name: sample
  serial: true
  plan:
  - get: tasks
  - get: image
    resource: pcf-automation
    params:
      globs: ["image-*.tar"]
      unpack: true

  - task: say hello
    image: image
    file: tasks/say-hello.yml
```

### `out`: not implemented

Not implemented. Returns an error

### Some common gotchas

#### Using glob patterns instead of regex patterns

We commonly see `product_version` patterns that look something like these:

```yaml
product_version: Go*          # Go Buildpack
#....
product_version: 1\.12\.*       # ERT
```

These superficially resemble globs, not regexes — but they are regexes. These will generally
work, but not because they are a glob. They work because the regex will also match.

For example, the first pattern, `Go*` will match "Go Buildpack 1.1.1". But it would also match
"Goooooooo" or "Go Tell It On A Mountain". The second pattern, `1\.12\.*`, will match "1.12.0".
But it will also match "1.12........." and "1.12.notanumber"

Instead, try patterns like:

```yaml
product_version: Go.*\d+\.\d+\.\d+  # Go Buildpack
#....
product_version: 1\.12\.\d+         # ERT
```

Note that [the regex syntax is Go's, which is slightly limited](https://github.com/google/re2/wiki/Syntax)
compared to PCRE and other popular syntaxes.

#### Using `check-resource` for sorted but non-sequential releases (eg. Buildpacks, Stemcells)

When doing a `check`, pivnet-resource defaults to using the server-provided order. This works
fine for simple cases where the response from the server is already in `semver` order. For example,
imagine this order from a product:

```
1.12.3
1.12.2
1.12.1
1.12.0
1.11.4
1.11.3
1.11.2
1.11.1
1.11.0
```

This list is in descending semver order. All the 1.12 patch releases are together, followed
by all the 1.11 patch releases and so on.

Some products do not group into `major` or `major.minor` groups in their responses. This is usually
because a product has multiple concurrent version releases. For example, stemcells typically
have multiple major versions available. When a CVE is announced that affects them, multiple
releases may occur at once, giving an order like:

```
9999.21
7777.19
9999.20
7777.18
```

In this example, the available versions for 9999 and 7777 are _sorted_ within the list, but not _sequential_.

To fix, use `sort_by: semver` in your resource definition.

Note that buildpack "versions" are actually a name and a version combined. You'll need to escape spaces
in your `check-resource` command for it to work properly. Eg:

```
fly -t pivnet check-resource \
  --resource pivnet-resource-bug-152616708/binary-buildpack \
  --from product_version:Binary\ 1.0.11#2017-03-23T13:57:51.214Z
```

In this example we escaped the space between "Binary" and "1.0.11".

## Integration environment

The Pivotal Network team maintains an integration environment at
`https://pivnet-integration.cfapps.io/`.

This environment is useful for teams to develop against, as changes to products
in this account are separated from the live account.

An example configuration for the integration environment might look like:

```yaml
resources:
- name: p-mysql
  type: pivnet-product-stemcell
  source:
    api_token: {{api-token}}
    product_slug: p-mysql
    stemcell_slug: stemcells-ubuntu-xenial
    endpoint: https://pivnet-integration.cfapps.io
```

## Developing

### Prerequisites

A valid install of golang is required - version 1.13.x is tested; earlier
versions may also work.

### Dependencies

We use [go modules](https://github.com/golang/go/wiki/Modules) for dependencies.


### Running the tests

Install the golint and ginkgo executables with:

```
go get -u golang.org/x/lint/golint
go get -u github.com/onsi/ginkgo/ginkgo
```

Run the tests with the following command:

```
ginkgo \
    -r \
    -race \
    -randomizeAllSpecs \
    -randomizeSuites \
    -keepGoing \
    -slowSpecThreshold="${SLOW_SPEC_THRESHOLD}"
```

And check for code formatting agains the Golang style guide with the following command:

```
golint ./...
```

### Contributing

Please make all pull requests to the `master` branch, and
[ensure the tests pass locally](https://github.com/shanman190/pivnet-product-stemcell-resource#running-the-tests).