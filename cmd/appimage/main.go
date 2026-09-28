// Command appimage packages the GUI without starting it.
package main

import (
	"fmt"
	"github.com/pieroproietti/penguins-gui/pkg/builder"
	"os"
)

func main() {
	path, err := builder.BuildAppImage(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Created", path)
}
