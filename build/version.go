package build

import (
	_ "embed"
)

var (
	//go:embed version.txt
	Version string
)
