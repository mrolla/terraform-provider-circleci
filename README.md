# CircleCI Terraform provider

[![Build Status](https://circleci.com/gh/mrolla/terraform-provider-circleci.svg?style=shield)](https://circleci.com/gh/mrolla/terraform-provider-circleci.svg?style=shield) [![Go Report Card](https://goreportcard.com/badge/github.com/mrolla/terraform-provider-circleci)](https://goreportcard.com/badge/github.com/mrolla/terraform-provider-circleci)

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

- [Terraform][terraform] 0.12.x (it has also been tested with version 0.11+)
- [Go][go] 1.11+ (to build the provider plugin)

## Using the provider

#### Download a release
Download the latest release for your OS from the [release page][release page]
and follow the instructions to [install third party plugins][third party plugins].

#### Build from sources
To build the project you can use `make all` which will:
- run the tests (`make test`)
- build the binary (`make build`)
- copy the binary to the [Terraform plugin directory][third party plugins] (`make install_plugin_locally`)

After placing it into your plugins directory, run `terraform init` to initialize it.

## Example:

```hcl
provider "circleci" {
  api_token    = "${file("circleci_token")}"
  vcs_type     = "github"
  organization = "my_org"
}

resource "circleci_environment_variable" "from_terraform" {
  project = "mySuperProject"
  name    = "from_terraform"
  value   = "the secret"
}
```

[install plugin]: https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin
[third party plugins]: https://www.terraform.io/docs/configuration/providers.html#third-party-plugins
[terraform]: https://www.terraform.io/downloads.html
[go]: https://golang.org/doc/install
[release page]: https://github.com/mrolla/terraform-provider-circleci/releases
