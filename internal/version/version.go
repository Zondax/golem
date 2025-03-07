package version

import _ "embed"

//go:generate bash get_version.sh
//go:embed version.txt
var GitVersion string
