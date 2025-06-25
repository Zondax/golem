# Common Docker Bake HCL variables and functions for reuse

variable "PLATFORMS" {
  default = "linux/amd64"
}

variable "EXTRA_TAGS" {
  default = ""
}

variable "BUILD_DATE" {
  default = "${timestamp()}"
}

variable "VERSION" {
  default = "latest"
}

# SBOM and Provenance configuration
variable "SBOM_FORMAT" {
  default = "spdx+json"
}

# Default target configuration with SBOM and provenance
target "docker-common" {
  args = {
    BUILD_DATE = "${BUILD_DATE}"
    VERSION = "${VERSION}"
  }
  context    = "."
  platforms = ["${PLATFORMS}"]
  attest = [
    "type=provenance,mode=max,builder-id=github-actions",
    "type=sbom"
  ]
}

function "generate_tags" {
  params = [base_name]
  result = concat(
    ["${base_name}:latest"],
    [for tag in split(" ", EXTRA_TAGS) : "${base_name}:${tag}" if tag != ""]
  )
}
