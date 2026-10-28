/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

// Command appimage packages the GUI without starting it.
package main

import (
	"fmt"
	"github.com/pieroproietti/penguins-gui/internal/builder"
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
