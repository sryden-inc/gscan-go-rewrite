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

    var allFileFlags map[string][]string

    for _, volume := range volumes {
        if volume.IsDir() {
            volumePath := filepath.Join(volumesDir, volume.Name())
            languagePercentages, fileFlags := analyzeFiles(volumePath, 1)
            if len(fileFlags) > 0 {
                printLanguagePercentages(volumePath, languagePercentages, fileFlags)
                allFileFlags = mergeMaps(allFileFlags, fileFlags)
            }
        }
    }

    printFlagSummary(allFileFlags)
}

func analyzeFiles(dirPath string, depth int) (map[string]float64, map[string][]string) {
    if depth > maxDepth {
        return nil, nil
    }

    languageCounts := make(map[string]int)
    totalFiles := 0
    fileFlags := make(map[string][]string)

    err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Ignore directories that start with "." or "?"
        if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "?")) {
            return filepath.SkipDir
        }

        if info.IsDir() && (info.Name() == "node_modules" || strings.HasPrefix(info.Name(), "node_modules"+string(os.PathSeparator))) {
            return filepath.SkipDir // skip node_modules directory and its subdirectories
        }

        if info.IsDir() && (info.Name() == "plugins" || strings.HasPrefix(info.Name(), "plugins"+string(os.PathSeparator))) {
            return filepath.SkipDir // skip node_modules directory and its subdirectories
        }

        if info.IsDir() && (info.Name() == "assets" || strings.HasPrefix(info.Name(), "assets"+string(os.PathSeparator))) {
            return filepath.SkipDir // skip node_modules directory and its subdirectories
        }

        if !info.IsDir() {
            ext := filepath.Ext(path)
            if ext == ".js" || ext == ".py" {
                content, err := readFileWithLimit(path, maxFileSize)
                if err != nil {
                    fmt.Printf("Error reading file %s: %v\n", path, err)
                    return nil
                }

                flags := checkFlags(string(content), path)
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

    for subDir, _ := range fileFlags {
        subDirPath := filepath.Dir(subDir)
        if subDirPath != dirPath {
            subdirLanguagePercentages, subdirFileFlags := analyzeFiles(subDirPath, depth+1)
            for ext, percentage := range subdirLanguagePercentages {
                languagePercentages[ext] += percentage
            }
            for path, flags := range subdirFileFlags {
                fileFlags[path] = flags
            }
        }
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

func checkFlags(content string, path string) []string {
    flags := []string{}
    if strings.Contains(content, "nezha") {
        flags = append(flags, "Nezha was detected")
    }
    if containsChinese(content) {
        flags = append(flags, "Contains Chinese characters")
    }
    if filepath.Ext(path) == ".sh" {
        flags = append(flags, "File ends with .sh")
    }
    if strings.Contains(content, "argo") || strings.Contains(content, "cloudflare") {
        flags = append(flags, "File contains 'argo' or 'cloudflare'")
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

func mergeMaps(dest, src map[string][]string) map[string][]string {
    if dest == nil {
        dest = make(map[string][]string)
    }
    for key, value := range src {
        dest[key] = value
    }
    return dest
}

func printFlagSummary(fileFlags map[string][]string) {
    if len(fileFlags) == 0 {
        fmt.Println("No flags found in any volume.")
        return
    }
    fmt.Println("Summary of flags found in all volumes:")
    for path, flags := range fileFlags {
        fmt.Printf("%s:\n", path)
        for _, flag := range flags {
            fmt.Printf("- %s\n", flag)
        }
    }
}
