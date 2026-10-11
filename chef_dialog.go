package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/pieroproietti/penguins-gui/internal/tools/chef"
)

// showCookDialog opens the interactive GUI recipe picker (apply.picker).
func showCookDialog(
	parent fyne.Window,
	chefPath string,
	chefRoot string,
	runCommand func(title string, executable string, args []string, confirmPrompt string, successMessage string),
	appendLog func(string),
) {
	recipesDir := filepath.Join(chefRoot, "recipes")
	rawRecipes, err := chef.FindRecipes(recipesDir, chefRoot)
	if err != nil {
		dialog.ShowError(err, parent)
		return
	}

	// Filtriamo via le categorie di basso livello (base, dm, de, costumes) per la GUI
	var recipes []chef.RecipeItem
	for _, r := range rawRecipes {
		cat := strings.ToLower(r.Category)
		if cat == "base" || cat == "dm" || cat == "de" || cat == "costumes" {
			continue
		}
		recipes = append(recipes, r)
	}

	if len(recipes) == 0 {
		dialog.ShowInformation("No Recipes Found", "No recipes found in "+chefRoot+". Run Update Recipes or 'chef get' to download them.", parent)
		return
	}

	var d *dialog.CustomDialog
	var selectedRecipe *chef.RecipeItem

	// Preview panel elements
	titleLabel := widget.NewLabelWithStyle("Select a recipe to view details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	categoryBadge := widget.NewLabel("")
	categoryBadge.TextStyle = fyne.TextStyle{Italic: true}
	descLabel := widget.NewLabel("")
	descLabel.Wrapping = fyne.TextWrapWord
	profilesLabel := widget.NewLabel("")
	profilesLabel.Wrapping = fyne.TextWrapWord
	profilesLabel.Hide()
	pathLabel := widget.NewLabel("")
	pathLabel.Wrapping = fyne.TextWrapBreak

	previewGrid := widget.NewTextGrid()
	previewScroll := container.NewScroll(previewGrid)
	previewScroll.SetMinSize(fyne.NewSize(460, 280))

	btnApply := widget.NewButtonWithIcon("Apply Recipe", theme.ConfirmIcon(), nil)
	btnApply.Importance = widget.HighImportance
	btnApply.Disable()

	btnDryRun := widget.NewButtonWithIcon("Dry Run", theme.VisibilityIcon(), nil)
	btnDryRun.Disable()

	statusLabel := widget.NewLabel("Select a recipe to begin.")

	selectRecipe := func(r *chef.RecipeItem) {
		if r == nil {
			selectedRecipe = nil
			titleLabel.SetText("Select a recipe to view details")
			categoryBadge.SetText("")
			descLabel.SetText("")
			profilesLabel.Hide()
			pathLabel.SetText("")
			previewGrid.SetText("")
			btnApply.Disable()
			btnDryRun.Disable()
			statusLabel.SetText("Select a recipe to begin.")
			return
		}

		selectedRecipe = r
		titleLabel.SetText(r.Name)
		categoryBadge.SetText("Category: " + r.Category)
		if r.Description != "" {
			descLabel.SetText(r.Description)
			descLabel.Show()
		} else {
			descLabel.Hide()
		}

		if len(r.Profiles) > 0 {
			profilesLabel.SetText("Supported distributions: " + strings.Join(r.Profiles, ", "))
			profilesLabel.Show()
		} else {
			profilesLabel.Hide()
		}

		pathLabel.SetText("File: " + r.RelPath)

		data, err := os.ReadFile(r.Path)
		if err == nil {
			previewGrid.SetText(string(data))
		} else {
			previewGrid.SetText(fmt.Sprintf("# Error reading file: %v", err))
		}
		previewScroll.ScrollToTop()
		btnApply.Enable()
		btnDryRun.Enable()
		statusLabel.SetText("Ready: " + r.Name)
	}

	selectCategory := func(catName string, count int, groupRecipes []chef.RecipeItem) {
		selectedRecipe = nil
		btnApply.Disable()
		btnDryRun.Disable()

		titleLabel.SetText("Category: " + catName + "/")
		categoryBadge.SetText(fmt.Sprintf("%d recipes available in this category", count))
		descLabel.SetText("Select an individual recipe to view its definition and apply it.")
		descLabel.Show()
		profilesLabel.Hide()
		pathLabel.SetText("")

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Category: %s/\n", catName))
		sb.WriteString(fmt.Sprintf("# Recipes found: %d\n#\n", count))
		for i, r := range groupRecipes {
			prefix := "├── "
			if i == len(groupRecipes)-1 {
				prefix = "└── "
			}
			sb.WriteString(fmt.Sprintf("#   %s%s\n", prefix, r.Name))
			if r.Description != "" {
				sb.WriteString(fmt.Sprintf("#       (%s)\n", r.Description))
			}
		}
		sb.WriteString("#\n# Select a recipe from the tree on the left to preview and apply.\n")
		previewGrid.SetText(sb.String())
		previewScroll.ScrollToTop()
		statusLabel.SetText("Category selected: " + catName)
	}

	// Build Tree data structures
	recipeMap := make(map[string]chef.RecipeItem)
	catChildren := make(map[string][]string)
	catMap := make(map[string][]chef.RecipeItem)
	groups := chef.GroupByCategory(recipes)

	var catUIDs []string
	for _, g := range groups {
		catID := "category:" + g.Name
		catUIDs = append(catUIDs, catID)
		catMap[g.Name] = g.Recipes

		var recUIDs []string
		for _, r := range g.Recipes {
			recID := fmt.Sprintf("recipe:%s/%s", r.Category, r.Name)
			recUIDs = append(recUIDs, recID)
			recipeMap[recID] = r
		}
		catChildren[catID] = recUIDs
	}
	catChildren[""] = catUIDs

	tree := widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return catChildren[uid]
		},
		func(uid widget.TreeNodeID) bool {
			return strings.HasPrefix(uid, "category:") || uid == ""
		},
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			if branch {
				icon.SetResource(theme.FolderIcon())
			}
			label := widget.NewLabel("")
			return container.NewHBox(icon, label)
		},
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			box := node.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)
			if branch {
				icon.SetResource(theme.FolderIcon())
				catName := strings.TrimPrefix(uid, "category:")
				count := len(catChildren[uid])
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.SetText(fmt.Sprintf("%s/ (%d)", catName, count))
			} else {
				icon.SetResource(theme.DocumentIcon())
				label.TextStyle = fyne.TextStyle{}
				if r, ok := recipeMap[uid]; ok {
					label.SetText(r.Name)
				} else {
					label.SetText(uid)
				}
			}
		},
	)

	tree.OnSelected = func(uid widget.TreeNodeID) {
		if strings.HasPrefix(uid, "category:") {
			catName := strings.TrimPrefix(uid, "category:")
			selectCategory(catName, len(catChildren[uid]), catMap[catName])
		} else if strings.HasPrefix(uid, "recipe:") {
			if r, ok := recipeMap[uid]; ok {
				selectRecipe(&r)
			}
		}
	}

	// Filter / Search functionality
	var filtered []chef.RecipeItem
	filteredList := widget.NewList(
		func() int { return len(filtered) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel(""))
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			if id < 0 || id >= len(filtered) {
				return
			}
			box := o.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			r := filtered[id]
			label.SetText(fmt.Sprintf("[%s] %s", r.Category, r.Name))
		},
	)
	filteredList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(filtered) {
			selectRecipe(&filtered[id])
		}
	}

	treeScroll := container.NewScroll(tree)
	listScroll := container.NewScroll(filteredList)
	listScroll.Hide()

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Filter recipes…")
	searchEntry.OnChanged = func(query string) {
		q := strings.TrimSpace(strings.ToLower(query))
		if q == "" {
			listScroll.Hide()
			treeScroll.Show()
			return
		}
		filtered = nil
		for _, r := range recipes {
			if strings.Contains(strings.ToLower(r.Name), q) ||
				strings.Contains(strings.ToLower(r.Category), q) ||
				strings.Contains(strings.ToLower(r.Description), q) {
				filtered = append(filtered, r)
			}
		}
		filteredList.Refresh()
		treeScroll.Hide()
		listScroll.Show()
	}

	btnRefresh := widget.NewButtonWithIcon("Update Recipes", theme.DownloadIcon(), func() {
		btnApply.Disable()
		btnDryRun.Disable()
		statusLabel.SetText("Updating ~/.chef recipes repository…")
		go func() {
			err := chef.DownloadChef(chefRoot, appendLog)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(fmt.Errorf("update failed: %w", err), parent)
					statusLabel.SetText("Update failed.")
				} else {
					statusLabel.SetText("Recipes updated successfully. Reopening picker…")
					d.Hide()
					showCookDialog(parent, chefPath, chefRoot, runCommand, appendLog)
				}
			})
		}()
	})

	btnApply.OnTapped = func() {
		if selectedRecipe == nil {
			return
		}
		r := *selectedRecipe
		d.Hide()
		runCommand(
			"Cook: "+r.Name,
			chefPath,
			[]string{"apply", r.Path},
			fmt.Sprintf("Apply recipe %q (%s)?\n\nFile: %s\n\nThis will install packages and configure the system according to the recipe.", r.Name, r.Category, r.Path),
			fmt.Sprintf("Recipe %q applied successfully.", r.Name),
		)
	}

	btnDryRun.OnTapped = func() {
		if selectedRecipe == nil {
			return
		}
		r := *selectedRecipe
		d.Hide()
		appendLog(fmt.Sprintf("$ chef apply --dry-run %s\n\n", r.Path))
		go func() {
			cmd := exec.Command(chefPath, "apply", "--dry-run", r.Path)
			out, err := cmd.CombinedOutput()
			fyne.Do(func() {
				appendLog(string(out) + "\n")
				if err != nil {
					dialog.ShowError(fmt.Errorf("dry-run failed: %w", err), parent)
				} else {
					dialog.ShowInformation("Dry-Run Completed", fmt.Sprintf("Dry-run preview for %q completed successfully.\nInspect the Operation log in the main window for details.", r.Name), parent)
				}
			})
		}()
	}

	btnClose := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		d.Hide()
	})

	// Layout construction
	leftContent := container.NewStack(treeScroll, listScroll)
	leftPanel := container.NewBorder(
		container.NewVBox(searchEntry),
		btnRefresh,
		nil,
		nil,
		leftContent,
	)

	metaBox := container.NewVBox(
		titleLabel,
		categoryBadge,
		descLabel,
		profilesLabel,
		pathLabel,
		widget.NewSeparator(),
	)
	rightPanel := container.NewBorder(
		metaBox,
		nil,
		nil,
		nil,
		previewScroll,
	)

	split := container.NewHSplit(leftPanel, rightPanel)
	split.SetOffset(0.35)

	const (
		compactW = float32(860)
		compactH = float32(600)
	)

	calcMax := func() (float32, float32) {
		w, h := compactW, compactH
		if parent != nil && parent.Canvas() != nil {
			cSize := parent.Canvas().Size()
			if cSize.Width-24 > compactW {
				w = cSize.Width - 24
			}
			if cSize.Height-24 > compactH {
				h = cSize.Height - 24
			}
		}
		return w, h
	}

	isMaximized := true
	var setCompact func()
	var setMaximized func()

	btnMin := widget.NewButtonWithIcon("", theme.WindowMinimizeIcon(), func() {
		if setCompact != nil {
			setCompact()
		}
	})
	btnMin.Importance = widget.LowImportance

	btnMax := widget.NewButtonWithIcon("", theme.ViewRestoreIcon(), func() {
		if isMaximized {
			if setCompact != nil {
				setCompact()
			}
		} else {
			if setMaximized != nil {
				setMaximized()
			}
		}
	})
	btnMax.Importance = widget.LowImportance

	btnCloseHeader := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() {
		if d != nil {
			d.Hide()
		}
	})
	btnCloseHeader.Importance = widget.LowImportance

	titleBox := container.NewHBox(
		widget.NewIcon(theme.ColorPaletteIcon()),
		widget.NewLabelWithStyle("Penguins Chef — Recipe Picker", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	topBar := container.NewVBox(
		container.NewBorder(
			nil,
			nil,
			titleBox,
			container.NewHBox(btnMin, btnMax, btnCloseHeader),
		),
		widget.NewSeparator(),
	)

	bottomBar := container.NewBorder(
		nil,
		nil,
		statusLabel,
		container.NewHBox(btnDryRun, btnApply, btnClose),
	)

	mainLayout := container.NewBorder(
		topBar,
		bottomBar,
		nil,
		nil,
		split,
	)

	d = dialog.NewCustomWithoutButtons("", mainLayout, parent)

	setCompact = func() {
		d.Resize(fyne.NewSize(compactW, compactH))
		btnMax.SetIcon(theme.WindowMaximizeIcon())
		isMaximized = false
	}

	setMaximized = func() {
		mw, mh := calcMax()
		d.Resize(fyne.NewSize(mw, mh))
		btnMax.SetIcon(theme.ViewRestoreIcon())
		isMaximized = true
	}

	initW, initH := calcMax()
	d.Resize(fyne.NewSize(initW, initH))
	d.Show()
	tree.OpenAllBranches()
}
