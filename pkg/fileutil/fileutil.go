// Package fileutil provides utilities for file and directory processing
package fileutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// FileInfo represents information about a file to be processed
type FileInfo struct {
	Path         string
	RelativePath string
	IsText       bool
	Size         int64
}

// IsTextFile determines if a file is likely to be a plain text file
func IsTextFile(path string) (bool, error) {
	// Check file extension first for quick filtering
	ext := strings.ToLower(filepath.Ext(path))

	// Known text file extensions
	textExtensions := []string{
		".txt", ".md", ".markdown", ".rst", ".adoc", ".asciidoc",
		".tex", ".latex", ".org", ".wiki", ".textile",
		".csv", ".tsv", ".json", ".xml", ".yaml", ".yml",
		".toml", ".ini", ".cfg", ".conf", ".config",
		".log", ".logs", ".out", ".err",
		".dockerfile", ".gitignore", ".gitattributes",
		".editorconfig", ".htaccess", ".robots",
		"", // files without extension
	}

	// Known binary extensions to exclude
	binaryExtensions := []string{
		".exe", ".bin", ".dll", ".so", ".dylib", ".a", ".o", ".obj",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".ico",
		".mp3", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".wav", ".ogg",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".tar", ".gz", ".bz2", ".xz", ".7z", ".rar",
		".deb", ".rpm", ".dmg", ".pkg", ".msi",
		".sqlite", ".db", ".sqlite3",
	}

	// Quick exclude for known binary extensions
	for _, binExt := range binaryExtensions {
		if ext == binExt {
			return false, nil
		}
	}

	// Quick include for known text extensions
	for _, txtExt := range textExtensions {
		if ext == txtExt {
			return true, nil
		}
	}

	// For unknown extensions, check file content
	return isTextFileByContent(path)
}

// isTextFileByContent checks if a file is text by examining its content
func isTextFileByContent(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = file.Close() // Ignore error in defer cleanup
	}()

	// Read first 512 bytes to check for binary content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return false, err
	}

	// Check if the content is valid UTF-8 and doesn't contain null bytes
	content := buffer[:n]

	// Null bytes are a strong indicator of binary content
	for _, b := range content {
		if b == 0 {
			return false, nil
		}
	}

	// Check if content is valid UTF-8
	if !utf8.Valid(content) {
		return false, nil
	}

	// Check for high ratio of control characters (excluding common ones)
	controlCount := 0
	for _, b := range content {
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			controlCount++
		}
	}

	// If more than 10% are control characters, likely binary
	if float64(controlCount)/float64(len(content)) > 0.1 {
		return false, nil
	}

	return true, nil
}

// FindTextFiles recursively finds all text files in a directory
func FindTextFiles(rootPath string) ([]FileInfo, error) {
	var files []FileInfo

	// Check if the path is a directory
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", rootPath, err)
	}

	if !info.IsDir() {
		// Single file
		isText, err := IsTextFile(rootPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check if file is text: %w", err)
		}

		if isText {
			files = append(files, FileInfo{
				Path:         rootPath,
				RelativePath: filepath.Base(rootPath),
				IsText:       true,
				Size:         info.Size(),
			})
		}

		return files, nil
	}

	// Directory - walk recursively
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log error but continue processing
			fmt.Fprintf(os.Stderr, "Warning: Error accessing %s: %v\n", path, err)
			return nil
		}

		// Skip common directories that should be ignored
		if d.IsDir() {
			dirName := strings.ToLower(d.Name())
			ignoredDirs := []string{
				".git", ".svn", ".hg", ".bzr",
				"node_modules", ".npm", ".yarn", "bower_components",
				".venv", "venv", "__pycache__", ".pytest_cache",
				"target", "build", "dist", "out", "bin",
				".idea", ".vscode", ".vs", ".settings",
				"vendor", ".gradle", ".m2",
				".cache", ".tmp", "tmp", "temp",
			}

			for _, ignored := range ignoredDirs {
				if dirName == ignored {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Check if it's a text file
		isText, err := IsTextFile(path)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: Error checking file type for %s: %v\n", path, err)
			return nil
		}

		if isText {
			info, err := d.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Error getting file info for %s: %v\n", path, err)
				return nil
			}

			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				relPath = path
			}

			files = append(files, FileInfo{
				Path:         path,
				RelativePath: relPath,
				IsText:       true,
				Size:         info.Size(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", rootPath, err)
	}

	return files, nil
}

// ReadFileContent reads the content of a file safely
func ReadFileContent(path string) (string, error) {
	// Check file size to avoid reading extremely large files
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	// Limit file size to 10MB for safety
	const maxFileSize = 10 * 1024 * 1024
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file %s is too large (%d bytes, max %d bytes)", path, info.Size(), maxFileSize)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return string(content), nil
}

// WriteFileContent writes content to a file safely
func WriteFileContent(path, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write with proper permissions
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// GetFileStats returns statistics about a set of files
func GetFileStats(files []FileInfo) map[string]interface{} {
	totalFiles := len(files)
	totalSize := int64(0)
	extCounts := make(map[string]int)

	for _, file := range files {
		totalSize += file.Size
		ext := strings.ToLower(filepath.Ext(file.Path))
		if ext == "" {
			ext = "(no extension)"
		}
		extCounts[ext]++
	}

	return map[string]interface{}{
		"total_files": totalFiles,
		"total_size":  totalSize,
		"extensions":  extCounts,
	}
}
