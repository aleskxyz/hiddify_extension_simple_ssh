package main

import (
	_ "github.com/aleskxyz/hiddify_extension_simple_ssh/hiddify_extension"

	"github.com/hiddify/hiddify-core/cmd"
)

func main() {
	cmd.StartExtension()
}
