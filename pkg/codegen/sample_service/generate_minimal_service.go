package sampleservice

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed sample_service_app/*
var minimalServiceFS embed.FS

// GenerateMinimalService generates a minimal service application structure
func GenerateMinimalService(serviceName string) error {
	sourceDir := "sample_service_app"
	cWd, err := os.Getwd()
	if err != nil {
		return err
	}

	servicePath := filepath.Join(cWd, serviceName)
	if _, err := os.Stat(servicePath); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", serviceName)
	}
	err = os.Mkdir(servicePath, 0755)
	if err != nil {
		return err
	}
	err = fs.WalkDir(minimalServiceFS, sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("error walking the path:", err)
		}
		if path == sourceDir { //skip the root directory
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		localPath := filepath.Join(servicePath, relPath)

		if d.IsDir() {
			return os.MkdirAll(localPath, 0755)
		}
		data, err := minimalServiceFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(localPath, data, 0644)
	})
	if err != nil {
		return err
	}
	err = os.Chdir(servicePath)
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	cmd := exec.Command("go", "mod", "init", serviceName)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}
	fmt.Println("Minimal service application structure generated successfully in ", servicePath)
	return nil
}
