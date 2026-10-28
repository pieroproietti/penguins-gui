/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/pieroproietti/penguins-gui/internal/system"
	"github.com/pieroproietti/penguins-gui/internal/ui"
)

//go:embed internal/builder/assets/penguins-gui.svg
var iconSVG []byte

func main() {
	if len(os.Args) > 1 && os.Args[1] == system.GRUBHelperArg {
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "Expected Eggs executable and ISO path")
			os.Exit(1)
		}
		if err := system.ConfigureEggsGRUB(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("ISO boot entry saved; previous marked Eggs entries replaced.")
		return
	}

	ui.Run(iconSVG)
}
