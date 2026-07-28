package main

import "github.com/wadefengx/wade/cmd"

// version is set via ldflags at build time
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
