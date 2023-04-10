package main

import _ "embed"

//go:generate bash scripts/get_version.sh
//go:embed scripts/version.txt
var GitVersion string

//go:generate bash scripts/get_revision.sh
//go:embed scripts/revision.txt
var GitRevision string
