# Terraform Provider

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
