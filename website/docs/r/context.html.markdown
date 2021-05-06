---
layout: "circleci"
page_title: "CircleCI: circleci_context"
sidebar_current: "docs-resource-circleci-context"
description: |-
  Manages CircleCI contexts.
---

# circleci_context

A CircleCI context is a named collection of environment variables that can be referenced in the configuration for workflows.

## Example Usage

Basic usage:

```hcl
resource "circleci_context" "build" {
  name  = "build"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the context.
* `organization` - (Optional) Organization where the context will be defined.

## Attributes Reference

* `id` - The ID of the context.

## Import

Contexts can be imported as `$organization/$context`, where "context" can be either a context name or ID. For example:

```shell
# name
terraform import circleci_context.build hashicorp/build

# id
terraform import circleci_context.build hashicorp/6d87b798-5edb-4d99-b424-ce73b43affb9
```
