package ui

import (
	"fmt"
	"path/filepath"

	"github.com/bnema/gart/internal/app"
)

func RunAddDotfileView(app *app.App, path string, dotfileName string) {
	path = app.ExpandHomeDir(path)
	cleanedPath := filepath.Clean(path)

	var err error
	if app.IsDir(path) {
		err = addDotfileDir(app, cleanedPath, dotfileName)
		if err != nil {
			fmt.Println(errorStyle.Render("Error:"), err)
			return
		}
	} else {
		err = addDotfileFile(app, cleanedPath, dotfileName)
		if err != nil {
			fmt.Println(errorStyle.Render("Error:"), err)
			return
		}
	}

	fmt.Println(boldStyle.Render("Adding dotfile:"), cleanedPath)

	if err != nil {
		fmt.Println(errorStyle.Render("Error:"), err)
		return
	}

	if err := app.GitCommitChanges("Add", dotfileName); err != nil {
		fmt.Println(errorStyle.Render("Error committing changes:"), err)
		return
	}

	fmt.Println(successStyle.Render("Dotfile added successfully!"))
}

func addDotfileDir(app *app.App, cleanedPath, dotfileName string) error {
	storePath := filepath.Join(app.StoragePath, dotfileName)

	if err := app.CopyDirectory(cleanedPath, storePath); err != nil {
		return fmt.Errorf("error copying directory: %w", err)
	}

	return updateConfig(app, cleanedPath, dotfileName)
}

func addDotfileFile(app *app.App, cleanedPath, dotfileName string) error {
	fileName := filepath.Base(cleanedPath)
	storePath := filepath.Join(app.StoragePath, fileName)

	if err := app.CopyFile(cleanedPath, storePath); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return updateConfig(app, cleanedPath, dotfileName)
}

func updateConfig(app *app.App, cleanedPath, dotfileName string) error {
	if err := app.UpdateConfig(dotfileName, cleanedPath); err != nil {
		return fmt.Errorf("error updating config: %w", err)
	}
	return nil
}
