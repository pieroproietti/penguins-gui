// Command package builds the native Linux archive without starting the GUI.
package main

import (
	"fmt"
	"os"

	"github.com/pieroproietti/penguins-gui/pkg/builder"
)

func main() {
	path, err := builder.Build(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Created", path)
}
