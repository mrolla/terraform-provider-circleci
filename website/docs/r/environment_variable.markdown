---
layout: "circleci"
page_title: "CircleCI: circleci_environment_variable"
sidebar_current: "docs-resource-circleci-environment-variable"
description: |-
  Manages a CircleCI project environment variable.
---

# circleci_environment_variable

A CircleCI environment variable is a key value pair added to a project that will be exposed to jobs.

## Example Usage

Basic usage:

```hcl
resource "circleci_environment_variable" "token" {
  name    = "TOKEN"
  value   = "secret"
  project = "project"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the environment variable.
* `value` - (Required) The value of the environment variable. A hash of this value will be stored in state in order to detect changes, but the plain text value will not be stored.
* `project` - (Required) The project that the environment variable will be added to.
* `organization` - (Optional) Organization where the project is defined.

## Import

Environment variables can be imported as `$organization.$project.$name`. For example:

```shell
terraform import circleci_environment_variable.token hashicorp.build.TOKEN
```
