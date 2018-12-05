# CircleCI terraform provider

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

- [Terraform][terraform] 0.10.x
- [Go][go] 1.11 (to build the provider plugin)

## Using the provider
If you're building the provider, follow the instructions to [install it as a plugin.][install plugin].
After placing it into your plugins directory,  run `terraform init` to initialize it.

[install plugin]: https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin
[terraform]: https://www.terraform.io/downloads.html
[go]: https://golang.org/doc/install


Example:
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

## Contribute

To build the project you can use `make all` which:
1. run the tests `make test`
2. build the binary `make build`
3. copy the binary to the terraform plugin directory (default $HOME/.terraform.d/plugins/)
 `make install_plugin_locally`