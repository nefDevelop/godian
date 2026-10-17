// Command godian es un generador de blog estático escrito en Go.
package main

import (
	"os"

	"github.com/nefdevelop/godian/internal/presentation/cli"
)

func main() {
	os.Exit(cli.Run(os.Args))
}
