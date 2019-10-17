module github.com/mrolla/terraform-provider-circleci

go 1.12

require (
	github.com/hashicorp/terraform v0.12.0
	github.com/jszwedko/go-circleci v0.2.0
	github.com/stretchr/testify v1.3.0
)

replace github.com/jszwedko/go-circleci v0.2.0 => github.com/tgermain/go-circleci v0.0.0-20181207123242-bfc5b3445bba
