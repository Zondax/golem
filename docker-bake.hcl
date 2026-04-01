variable "BASE_NAME" {
  default = "golem"
}

variable "REGISTRY" {
  default = "zondax"
}

variable "GIT_SHORT_HASH" {
  default = ""
}

variable "GIT_BRANCH" {
  default = ""
}

variable "GIT_TAG" {
  default = ""
}

variable "GIT_COMMIT_TIMESTAMP" {
  default = ""
}

function "flex_tags" {
  params = [name]
  result = compact([
    "${REGISTRY}/${name}:latest",
    notequal(GIT_SHORT_HASH, "") ? "${REGISTRY}/${name}:${GIT_SHORT_HASH}" : "",
    notequal(GIT_BRANCH, "") ? "${REGISTRY}/${name}:${GIT_BRANCH}" : "",
    notequal(GIT_COMMIT_TIMESTAMP, "") ? "${REGISTRY}/${name}:T${GIT_COMMIT_TIMESTAMP}" : "",
    notequal(GIT_TAG, "") ? "${REGISTRY}/${name}:${GIT_TAG}" : "",
  ])
}

group "default" {
  targets = ["production"]
}

target "builder" {
  context    = "."
  dockerfile = "build/Dockerfile"
  target     = "builder"
  tags       = ["${REGISTRY}/${BASE_NAME}:builder"]
}

target "production" {
  context    = "."
  dockerfile = "build/Dockerfile"
  tags       = flex_tags(BASE_NAME)
  platforms  = ["linux/amd64"]
}
