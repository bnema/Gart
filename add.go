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

var ignoredDirs = []string{
	".svelte-kit", "local", ".cache", ".git", ".github", ".store", "ollama", "spotify",
	".vscode", ".idea", "node_modules", "vendor", "target",
	"build", "dist", "out", "bin", "obj", "logs", "tmp",
	"backup", "backups", "cache", "caches", "temp", "temps",
	"tmps", "tmpfs", "tempfs", ".var", ".store", "yay", "src", "pkg", "bin", "lib", "include", "share", "local", "opt",
	"pacman", "snap", "flatpak", "paru", ".npm", ".pnpm", ".yarn", ".cargo", ".steam",
}

func isIgnoredDir(dirPath string) bool {
	for _, dir := range ignoredDirs {
		if strings.Contains(dirPath, dir) {
			return true
		}
	}
	return false
}

func (app *App) addDotfile(path, name string) {
	// If the path starts with a tilde, replace it with the user's home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Erreur lors de l'obtention du répertoire personnel : %v\n", err)
			return
		}
		path = home + path[1:]
	}
	if strings.HasSuffix(path, "*") {
		directoryPath := strings.TrimSuffix(path, "*")

		// Create a channel to send directory paths to
		dirChan := make(chan string, 1000)

		// Create a WaitGroup to wait for all the worker goroutines to finish
		var wg sync.WaitGroup

		// Get the number of CPU cores
		numWorkers := runtime.NumCPU()

		// Démarrez les goroutines des travailleurs
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for dirPath := range dirChan {
					// Check if the directory is ignored or is the same as the destination directory
					if isIgnoredDir(dirPath) || strings.HasPrefix(dirPath, app.StorePath) {
						fmt.Printf("Répertoire ignoré : %s\n", dirPath)
						continue
					}
					destPath := filepath.Join(app.StorePath, name, strings.TrimPrefix(dirPath, filepath.Dir(directoryPath)))
					err := system.CopyDirectory(dirPath, destPath)
					if err != nil {
						fmt.Printf("Erreur lors de la copie du répertoire : %v\n", err)
					}
				}
			}()
		}

		// Parcourez l'arborescence des répertoires et envoyez les chemins des répertoires au canal
		err := filepath.Walk(directoryPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
				dirChan <- filePath
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Erreur lors de la traversée du répertoire : %v\n", err)
			return
		}

		// Fermez le canal pour signaler aux travailleurs de s'arrêter
		close(dirChan)

		// Attendez que toutes les goroutines des travailleurs se terminent
		wg.Wait()
	} else {
		cleanedPath := filepath.Clean(path)

		storePath := filepath.Join(app.StorePath, name)
		err := system.CopyDirectory(cleanedPath, storePath)
		if err != nil {
			fmt.Printf("Erreur lors de la copie du répertoire : %v\n", err)
			return
		}

		app.ListModel.dotfiles[name] = cleanedPath
		err = config.SaveConfig(app.ConfigFilePath, app.ListModel.dotfiles)
		if err != nil {
			fmt.Printf("Erreur lors de l'enregistrement de la configuration : %v\n", err)
			return
		}

		fmt.Printf("Dotfile ajouté : %s\n", name)
	}
}
