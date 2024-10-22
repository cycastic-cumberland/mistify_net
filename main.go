package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func getAllFiles(root string, filter func(string) bool, filesChan chan<- string, wg *sync.WaitGroup) {
	defer func() {
		close(filesChan)
		wg.Done()
	}()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filter != nil && !filter(path) {
			return nil
		}
		if !info.IsDir() {
			filesChan <- path
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", root, err)
	}
}

func isWhiteCharacter(c uint8) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
}

func isXmlDoc(line string) bool {
	for i := 0; i < len(line); i++ {
		c := line[i]
		if isWhiteCharacter(c) {
			continue
		}
		if c == '/' {
			if len(line)-i < 3 {
				return false
			}

			if line[i+1] != '/' || line[i+2] != '/' {
				return false
			}

			return true
		}
	}

	return false
}

func removeXmlDocs(path string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Found file:", path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	content := string(bytes)
	lines := strings.Split(content, "\n")
	cleansedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if isXmlDoc(line) {
			continue
		}
		cleansedLines = append(cleansedLines, line)
	}
	finalLines := strings.Join(cleansedLines, "\n")
	err = os.WriteFile(path, []byte(finalLines), 0644)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <path>\n", os.Args[0])
		return
	}
	root := os.Args[1]
	filesChan := make(chan string)
	var wg sync.WaitGroup

	wg.Add(1)
	go getAllFiles(root, func(s string) bool {
		return filepath.Ext(s) == ".cs"
	}, filesChan, &wg)

	// Receive file paths from the channel
	for path := range filesChan {
		wg.Add(1)
		go removeXmlDocs(path, &wg)
	}

	wg.Wait()
}
