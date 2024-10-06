package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/koeng101/dnadesign/lib/seqhash"
	"gopkg.in/yaml.v2"
)

// Embed the entire parts directory
//
//go:embed parts
var embeddedFiles embed.FS

// Part represents a single part part.
type Part struct {
	Seqhash     string   `yaml:"seqhash" json:"seqhash"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Tags        []string `yaml:"tags" json:"tags"`
	Prefix      string   `yaml:"prefix" json:"prefix"`
	Suffix      string   `yaml:"suffix" json:"suffix"`
	Sequence    string   `yaml:"sequence" json:"sequence"`
}

func main() {
	// Use fs.WalkDir to walk through embedded directory
	partMap := make(map[string]Part)
	err := fs.WalkDir(embeddedFiles, "parts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error walking through embedded directory: %v\n", err)
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			data, err := embeddedFiles.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading embedded file %s: %v\n", path, err)
				return err
			}

			var contents map[string]Part
			err = yaml.Unmarshal(data, &contents)
			if err != nil {
				fmt.Printf("Error unmarshalling YAML from embedded file %s: %v\n", path, err)
				return err
			}

			for name, part := range contents {
				if part.Prefix == "" || part.Suffix == "" || part.Sequence == "" {
					continue
				}
				sq, err := seqhash.EncodeHash2(seqhash.Hash2Fragment(strings.ToUpper(part.Prefix+part.Sequence+part.Suffix), 4, 4))
				if err != nil {
					fmt.Printf("Error seqhashing: %v\n", err)
					return err
				}
				partMap[sq] = Part{Name: name, Seqhash: sq, Description: part.Description, Tags: part.Tags, Prefix: strings.ToUpper(part.Prefix), Suffix: strings.ToUpper(part.Suffix), Sequence: strings.ToUpper(part.Sequence)}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the embedded file system: %v\n", err)
	}

	/*
		Build parts directory
	*/

	directory := "build/parts"

	// Ensure the directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, 0755)
	}

	// Clear all files in the directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}
	for _, file := range files {
		err := os.RemoveAll(filepath.Join(directory, file.Name()))
		if err != nil {
			fmt.Println("Error removing file:", err)
			return
		}
	}
	// Serialize and write files for each part
	for _, part := range partMap {
		fmt.Println(part)
		jsonFileName := filepath.Join(directory, part.Seqhash+".json")
		yamlFileName := filepath.Join(directory, part.Seqhash+".yaml")

		// Marshal part to JSON
		jsonData, err := json.MarshalIndent(part, "", "    ")
		if err != nil {
			fmt.Println("Error marshaling to JSON:", err)
			continue
		}

		// Marshal part to YAML
		yamlData, err := yaml.Marshal(part)
		if err != nil {
			fmt.Println("Error marshaling to YAML:", err)
			continue
		}

		// Write JSON file
		if err := ioutil.WriteFile(jsonFileName, jsonData, 0644); err != nil {
			fmt.Println("Error writing JSON file:", err)
			continue
		}

		// Write YAML file
		if err := ioutil.WriteFile(yamlFileName, yamlData, 0644); err != nil {
			fmt.Println("Error writing YAML file:", err)
			continue
		}
	}
}
