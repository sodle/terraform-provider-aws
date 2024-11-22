# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_appmesh_mesh" "test" {
  name = aws_appmesh_mesh.test.name
}

resource "aws_appmesh_mesh" "test" {
  name = var.rName

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
