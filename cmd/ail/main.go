// Command ail is the executable entry point for the AIL toolchain.
package main

import (
	"os"

	"github.com/gtihq/ail/packages/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
