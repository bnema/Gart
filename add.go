package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/bnema/Gart/config"
	"github.com/bnema/Gart/system"
)

func (app *App) addDotfile(path, name string) {
	// If the path starts with ~, replace it with the user's home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		path = home + path[1:]
	}

	if strings.HasSuffix(path, "*") {
		directoryPath := strings.TrimSuffix(path, "*")

		// Use filepath.Glob to get the list of directories matching the pattern
		dirs, err := filepath.Glob(filepath.Join(directoryPath, ".*"))
		if err != nil {
			fmt.Printf("Error using filepath.Glob: %v\n", err)
			return
		}

		// Create a wait group to wait for all worker goroutines to finish
		var wg sync.WaitGroup

		// Determine the number of worker goroutines based on available CPU cores
		numWorkers := runtime.NumCPU()

		// Create a buffered channel to hold the directory paths
		dirChan := make(chan string, len(dirs))

		// Send the directory paths to the channel
		for _, dir := range dirs {
			if (filepath.Base(dir) == ".config" && strings.Contains(dir, "gart")) || filepath.Base(dir) == ".local" {
				fmt.Printf("Ignored directory: %s\n", dir)
				continue
			}
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				dirChan <- dir
			}
		}
		close(dirChan)

		// Start the worker goroutines
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for dirPath := range dirChan {
					relPath, _ := filepath.Rel(directoryPath, dirPath)
					destPath := filepath.Join(app.StorePath, name, relPath)
					system.CopyDirectory(dirPath, destPath)
				}
			}()
		}

		// Wait for all worker goroutines to finish
		wg.Wait()
	} else {
		cleanedPath := filepath.Clean(path)

		storePath := filepath.Join(app.StorePath, name)
		err := system.CopyDirectory(cleanedPath, storePath)
		if err != nil {
			fmt.Printf("Error copying directory: %v\n", err)
			return
		}

		app.ListModel.dotfiles[name] = cleanedPath
		err = config.SaveConfig(app.ConfigFilePath, app.ListModel.dotfiles)
		if err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			return
		}

		// TODO: Create a state with git with the date

		fmt.Printf("Dotfile added: %s\n", name)
	}
}
