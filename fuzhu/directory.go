package fuzhu

import (
	"os"
	"path/filepath"
)

type DirectoryManager struct {
}

func NewDirectoryManager() *DirectoryManager {
	return &DirectoryManager{}
}

func (dm *DirectoryManager) CreateDirectory(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}
func (dm *DirectoryManager) Delete(path string) error {
	return os.RemoveAll(path)
}
func (dm *DirectoryManager) Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
func (dm *DirectoryManager) GetDirectories(path string) ([]string, error) {
	var dirs []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(path, entry.Name()))
		}
	}
	return dirs, nil
}

func (dm *DirectoryManager) GetDirectoriesAllDirectories(path string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	return dirs, err
}

func (dm *DirectoryManager) GetDirectoriesSearchPattern(path string, searchPattern string) ([]string, error) {
	var dirs []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			match, err := filepath.Match(searchPattern, name)
			if err != nil {
				return nil, err
			}
			if match {
				dirs = append(dirs, filepath.Join(path, name))
			}
		}
	}
	return dirs, nil
}

func (dm *DirectoryManager) GetDirectoriesSearchPatternAllDirectories(path string, searchPattern string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := filepath.Base(path)
			match, err := filepath.Match(searchPattern, name)
			if err != nil {
				return err
			}
			if match {
				dirs = append(dirs, path)
			}
		}
		return nil
	})
	return dirs, err
}

func (dm *DirectoryManager) GetFiles(path string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	return files, nil
}

func (dm *DirectoryManager) GetFilesAllDirectories(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (dm *DirectoryManager) GetFilesSearchPattern(path string, searchPattern string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			match, err := filepath.Match(searchPattern, name)
			if err != nil {
				return nil, err
			}
			if match {
				files = append(files, filepath.Join(path, name))
			}
		}
	}
	return files, nil
}

func (dm *DirectoryManager) GetFilesSearchPatternAllDirectories(path string, searchPattern string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			name := filepath.Base(path)
			match, err := filepath.Match(searchPattern, name)
			if err != nil {
				return err
			}
			if match {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}
