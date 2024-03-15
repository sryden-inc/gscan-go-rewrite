package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const maxFileSize = 1024 * 1024 * 10 // 10 MB

func main() {
    volumesDir := "/var/lib/pterodactyl/volumes/"
    volumes, err := ioutil.ReadDir(volumesDir)
    if err != nil {
        fmt.Println("Error reading volumes directory:", err)
        return
    }

    for _, volume := range volumes {
        if volume.IsDir() {
            volumePath := filepath.Join(volumesDir, volume.Name())
            languagePercentages, fileFlags := analyzeFiles(volumePath)
            if len(fileFlags) > 0 {
                printLanguagePercentages(volumePath, languagePercentages, fileFlags)
            }
        }
    }
}

func analyzeFiles(dirPath string) (map[string]float64, map[string][]string) {
    languageCounts := make(map[string]int)
    totalFiles := 0
    fileFlags := make(map[string][]string)

    err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() {
            ext := filepath.Ext(path)
            if ext == ".js" || ext == ".py" {
                content, err := readFileWithLimit(path, maxFileSize)
                if err != nil {
                    fmt.Printf("Error reading file %s: %v\n", path, err)
                    return nil
                }

                flags := checkFlags(string(content))
                if len(flags) > 0 {
                    fileFlags[path] = flags
                }
            }
            languageCounts[ext]++
            totalFiles++
        }

        return nil
    })

    if err != nil {
        fmt.Printf("Error walking directory %s: %v\n", dirPath, err)
        return nil, nil
    }

    languagePercentages := make(map[string]float64)
    for ext, count := range languageCounts {
        percentage := float64(count) / float64(totalFiles) * 100.0
        languagePercentages[ext] = percentage
    }

    return languagePercentages, fileFlags
}

func readFileWithLimit(path string, limit int64) ([]byte, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    fi, err := file.Stat()
    if err != nil {
        return nil, err
    }

    if fi.Size() > limit {
        return nil, fmt.Errorf("file %s is too large (size: %d bytes, limit: %d bytes)", path, fi.Size(), limit)
    }

    return ioutil.ReadAll(file)
}

func checkFlags(content string) []string {
    flags := []string{}
    if strings.Contains(content, "nezha") {
        flags = append(flags, "Nezha was detected")
    }
    if containsChinese(content) {
        flags = append(flags, "Contains Chinese characters")
    }
    if strings.Contains(content, "root") {
        flags = append(flags, "Contains references to 'root'")
    }
    return flags
}

func containsChinese(s string) bool {
    for _, r := range s {
        if unicode.Is(unicode.Scripts["Han"], r) {
            return true
        }
    }
    return false
}

func printLanguagePercentages(dirPath string, languagePercentages map[string]float64, fileFlags map[string][]string) {
    fmt.Printf("Directory: %s\n\nLanguages:\n", dirPath)
    for ext, percentage := range languagePercentages {
        language := strings.TrimPrefix(ext, ".")
        fmt.Printf("* %.0f%% %s\n", percentage, language)
    }

    fmt.Println("\nFlags found in files:")
    for path, flags := range fileFlags {
        fmt.Printf("%s:\n", path)
        for _, flag := range flags {
            fmt.Printf("- %s\n", flag)
        }
    }

    fmt.Println()
}