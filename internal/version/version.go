package version

import _ "embed"

//go:generate bash get_version.sh
//go:embed version.txt
var GitVersion string

//go:generate bash get_revision.sh
//go:embed revision.txt
var GitRevision string
