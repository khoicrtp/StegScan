package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

var FILENAME = ""
var OUTPUT_DIR = ""

// MagicHeader defines a file format's magic header and extraction function.
type MagicHeader struct {
	FileType    string
	MagicBytes  []byte
	ExtractFunc func(data []byte, outputDir string) error
}

func readMagicHeaders(filename string) ([]MagicHeader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	headers := []MagicHeader{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format in file signatures: %s", line)
		}

		fileType := parts[0]
		hexSignature := parts[1]
		fmt.Println("fileType:", fileType, "hexSignature:", hexSignature)
		magicBytes, err := hex.DecodeString(hexSignature)
		if err != nil {
			return nil, err
		}

		header := MagicHeader{
			FileType:   fileType,
			MagicBytes: magicBytes,
		}

		switch fileType {
		case "PNG":
			header.ExtractFunc = func(data []byte, outputDir string) error {
				reader := bytes.NewReader(data)
				image, _, err := image.Decode(reader)
				if err != nil {
					return err
				}

				outputFile, err := os.Create(filepath.Join(outputDir, FILENAME+".png"))
				if err != nil {
					return err
				}
				defer outputFile.Close()

				return png.Encode(outputFile, image)
			}
		case "GIF":
			header.ExtractFunc = func(data []byte, outputDir string) error {
				reader := bytes.NewReader(data)
				image, _, err := image.Decode(reader)
				if err != nil {
					return err
				}

				outputFile, err := os.Create(filepath.Join(outputDir, FILENAME+".gif"))
				if err != nil {
					return err
				}
				defer outputFile.Close()

				return gif.Encode(outputFile, image, nil)
			}
		// Add more cases for other file types as needed
		default:
			header.ExtractFunc = func(data []byte, outputDir string) error {
				outputFile, err := os.Create(filepath.Join(outputDir, FILENAME+"."+strings.ToLower(fileType)))
				if err != nil {
					return err
				}
				defer outputFile.Close()

				_, err = outputFile.Write(data)
				return err
			}
		}

		headers = append(headers, header)
	}

	return headers, nil
}

func extractEmbeddedFiles(inputFilePath string, outputDir string, magicHeaders []MagicHeader) error {
	// Read the entire input file
	data, err := os.ReadFile(inputFilePath)
	if err != nil {
		return err
	}

	// Iterate through each magic header
	for _, header := range magicHeaders {
		if bytes.Contains(data, header.MagicBytes) {
			fmt.Printf("Extracting %s...\n", header.FileType)
			if err := header.ExtractFunc(data, outputDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	var inputFilePath string
	if len(os.Args) == 2 {
		inputFilePath = os.Args[1]
		OUTPUT_DIR = "./output"
	} else {
		inputFilePath = os.Args[1]
		OUTPUT_DIR = os.Args[2]
	}

	FILENAME = strings.Split(inputFilePath, "/")[len(strings.Split(inputFilePath, "/"))-1]

	magicHeaders, err := readMagicHeaders("type.txt")
	if err != nil {
		fmt.Printf("Error reading file signatures: %v\n", err)
		return
	}
	print(magicHeaders)

	err = extractEmbeddedFiles(inputFilePath, OUTPUT_DIR, magicHeaders)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
