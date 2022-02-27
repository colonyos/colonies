package main

import (
	"github.com/colonyos/colonies/internal/cli"
	"github.com/colonyos/colonies/pkg/build"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

func main() {
	build.BuildVersion = BuildVersion
	build.BuildTime = BuildTime
	cli.Execute()
}
