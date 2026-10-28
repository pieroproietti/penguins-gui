/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

// Command package builds the native Linux archive without starting the GUI.
package main

import (
	"fmt"
	"os"

	"github.com/pieroproietti/penguins-gui/internal/builder"
)

func main() {
	path, err := builder.Build(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Created", path)
}
