package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func prepareImageDirectory(tempDirectory string) string {
	dir := path.Join(tempDirectory, fmt.Sprintf("%d", time.Now().UnixMilli()))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Errorf("error creating directory: %w", err))
	}
	return dir
}

func getLatestDrawing(dir string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(fmt.Errorf("error reading directory: %w", err))
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
			fileNames = append(fileNames, file.Name())
		}
	}

	if len(fileNames) == 0 {
		return ""
	}

	sort.Strings(fileNames)
	lastFileName := fileNames[len(fileNames)-1]
	return filepath.Join(dir, lastFileName)
}

func getImageB64(imagePath string) string {
	imageFile, err := os.Open(imagePath)
	if err != nil {
		panic(fmt.Errorf("error opening image: %w", err))
	}
	defer func(file *os.File) {
		e := file.Close()
		if e != nil {
			panic(fmt.Errorf("error closing image: %w", e))
		}
	}(imageFile)

	var imageData []byte
	if imageData, err = io.ReadAll(imageFile); err != nil {
		panic(fmt.Errorf("error reading file: %w", err))
	}

	return base64.StdEncoding.EncodeToString(imageData)
}
