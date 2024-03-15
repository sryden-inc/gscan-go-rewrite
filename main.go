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
const maxDepth = 3                   // maximum depth of directory traversal

func main() {
    volumesDir := "/var/lib/pterodactyl/volumes/"
    volumes, err := ioutil.ReadDir(volumesDir)
    if err != nil {
        fmt.Println("Error reading volumes directory:", err)
        return
    }

    flaggedFolders := make(map[string]bool)

    for _, volume := range volumes {
        if volume.IsDir() {
            volumePath := filepath.Join(volumesDir, volume.Name())
            languagePercentages, fileFlags, folderFlags := analyzeFiles(volumePath, 1)
            if len(fileFlags) > 0 || len(folderFlags) > 0 {
                printLanguagePercentages(volumePath, languagePercentages, fileFlags)
                for folder := range folderFlags {
                    flaggedFolders[folder] = true
                }
            }
        }
    }

    printFlaggedFoldersSummary(flaggedFolders)
}

func analyzeFiles(dirPath string, depth int) (map[string]float64, map[string][]string, map[string]bool) {
    if depth > maxDepth {
        return nil, nil, nil
    }

    languageCounts := make(map[string]int)
    totalFiles := 0
    fileFlags := make(map[string][]string)
    folderFlags := make(map[string]bool)

    err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            if info.Name() == "node_modules" || strings.HasPrefix(info.Name(), "node_modules"+string(os.PathSeparator)) ||
                strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "/assets/") {
                folderFlags[path] = true
                return filepath.SkipDir // skip node_modules, .*, and /assets/ directories and their subdirectories
            }
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
        return nil, nil, nil
    }

    languagePercentages := make(map[string]float64)
    for ext, count := range languageCounts {
        percentage := float64(count) / float64(totalFiles) * 100.0
        languagePercentages[ext] = percentage
    }

    for subDir, _ := range fileFlags {
        subDirPath := filepath.Dir(subDir)
        if subDirPath != dirPath {
            subdirLanguagePercentages, subdirFileFlags, subdirFolderFlags := analyzeFiles(subDirPath, depth+1)
            for ext, percentage := range subdirLanguagePercentages {
                languagePercentages[ext] += percentage
            }
            for path, flags := range subdirFileFlags {
                fileFlags[path] = flags
            }
            for folder := range subdirFolderFlags {
                folderFlags[folder] = true
            }
        }
    }

    return languagePercentages, fileFlags, folderFlags
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

func printFlaggedFoldersSummary(flaggedFolders map[string]bool) {
    if len(flaggedFolders) > 0 {
        fmt.Println("\nThese are all of the flagged volumes:")
        for folder := range flaggedFolders {
            fmt.Println(folder)
        }
    }
}
