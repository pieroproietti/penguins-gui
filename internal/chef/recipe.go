/*
 * penguins-gui - Copyright (C) 2026-2026 Piero Proietti
 * Author: Piero Proietti <piero.proietti@gmail.com>
 * https://penguins-eggs.net
 *
 * This file is part of penguins-gui.
 * Released under the GNU GPL v3 license.
 */

package chef

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// RecipeItem represents an individual recipe discovered on disk.
type RecipeItem struct {
	Name        string   `yaml:"name"`
	Category    string   `yaml:"-"`
	Path        string   `yaml:"-"`
	RelPath     string   `yaml:"-"`
	Description string   `yaml:"description"`
	Hostname    string   `yaml:"hostname"`
	Include     []string `yaml:"include"`
	Profiles    []string `yaml:"-"`
	Sysroot     string   `yaml:"sysroot"`
}

// CategoryGroup groups recipes under a category name.
type CategoryGroup struct {
	Name    string
	Recipes []RecipeItem
}

type rawRecipeYAML struct {
	Version     int                       `yaml:"version"`
	Name        string                    `yaml:"name"`
	Description string                    `yaml:"description"`
	Hostname    string                    `yaml:"hostname"`
	Include     []string                  `yaml:"include"`
	Profiles    map[string]map[string]any `yaml:"profiles"`
	Sysroot     string                    `yaml:"sysroot"`
}

// FindRecipes scans the specified directories recursively for .yaml/.yml recipes.
func FindRecipes(dirs ...string) ([]RecipeItem, error) {
	seen := make(map[string]bool)
	var recipes []RecipeItem

	for _, dir := range dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}

		stat, err := os.Stat(absDir)
		if err != nil || !stat.IsDir() {
			continue
		}

		err = filepath.Walk(absDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}

			if info.IsDir() {
				name := info.Name()
				if strings.HasPrefix(name, ".") || name == "sysroot" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}

			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}

			rel, err := filepath.Rel(absDir, path)
			if err != nil {
				rel = info.Name()
			}
			slashRel := filepath.ToSlash(rel)

			parts := strings.Split(slashRel, "/")
			category := "general"
			if len(parts) > 1 {
				category = parts[0]
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			var parsed rawRecipeYAML
			if err := yaml.Unmarshal(data, &parsed); err != nil {
				return nil
			}

			recipeName := parsed.Name
			if recipeName == "" {
				recipeName = strings.TrimSuffix(info.Name(), ext)
			}

			var profiles []string
			for p := range parsed.Profiles {
				profiles = append(profiles, p)
			}
			sort.Strings(profiles)

			dedupKey := category + "/" + recipeName
			if seen[dedupKey] {
				return nil
			}
			seen[dedupKey] = true

			recipes = append(recipes, RecipeItem{
				Name:        recipeName,
				Category:    category,
				Path:        path,
				RelPath:     slashRel,
				Description: strings.TrimSpace(parsed.Description),
				Hostname:    parsed.Hostname,
				Include:     parsed.Include,
				Profiles:    profiles,
				Sysroot:     parsed.Sysroot,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(recipes, func(i, j int) bool {
		if recipes[i].Category != recipes[j].Category {
			return recipes[i].Category < recipes[j].Category
		}
		return recipes[i].Name < recipes[j].Name
	})

	return recipes, nil
}

// GroupByCategory groups recipes by category into sorted CategoryGroup slices.
func GroupByCategory(recipes []RecipeItem) []CategoryGroup {
	catMap := make(map[string][]RecipeItem)
	for _, r := range recipes {
		cat := r.Category
		if cat == "" {
			cat = "general"
		}
		catMap[cat] = append(catMap[cat], r)
	}

	var catNames []string
	for c := range catMap {
		catNames = append(catNames, c)
	}
	sort.Strings(catNames)

	var groups []CategoryGroup
	for _, c := range catNames {
		items := catMap[c]
		sort.Slice(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})
		groups = append(groups, CategoryGroup{
			Name:    c,
			Recipes: items,
		})
	}
	return groups
}
