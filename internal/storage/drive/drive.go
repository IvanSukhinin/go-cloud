package drive

import (
	"bytes"
	"cloud/internal/config"
	"cloud/internal/storage"
	"errors"
	"fmt"
	"github.com/djherbis/times"
	"os"
	"sync"
)

type Storage struct {
	tmpPath       string
	completedPath string
	mu            sync.Mutex
}

type Image struct {
	Name      string
	CreatedAt string
	UpdatedAt string
}

// New init storage.
func New(cfg config.StorageConfig) (*Storage, error) {
	tmpPath := cfg.TmpPath
	info, err := os.Stat(tmpPath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("tmp storage must be directory")
	}

	completedPath := cfg.CompletedPath
	info, err = os.Stat(completedPath)
	if os.IsNotExist(err) {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("completed storage must be directory")
	}

	return &Storage{
		tmpPath:       tmpPath,
		completedPath: completedPath,
	}, nil
}

// Save saves image on disk.
func (s *Storage) Save(filename string, buf bytes.Buffer) error {
	const fn = "drive.Save"

	file, err := s.createFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = buf.WriteTo(file)
	if err != nil {
		err := os.Remove(s.tmpPath + filename)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}
		return fmt.Errorf("%s: cannot write buf to file: %w", fn, err)
	}

	err = s.successUpload(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// createFile checks if the file exists and saves it thread safe.
func (s *Storage) createFile(filename string) (*os.File, error) {
	const fn = "drive.createFile"

	s.mu.Lock()
	defer s.mu.Unlock()

	// check if file exists
	isExist, err := s.FileExists(filename)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, fmt.Errorf("%s: %w", fn, storage.ErrFileExists)
	}

	// create file
	path := fmt.Sprintf("%s/%s", s.tmpPath, filename)
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot create image file: %w", fn, err)
	}
	return file, nil
}

// successUpload move file to completed directory
func (s *Storage) successUpload(filename string) error {
	const fn = "drive.successUpload"
	err := os.Rename(s.tmpPath+filename, s.completedPath+filename)
	if err != nil {
		return fmt.Errorf("%v: %w", fn, err)
	}
	return nil
}

// List returns all images.
func (s *Storage) List() ([]Image, error) {
	const fn = "drive.List"

	entries, err := os.ReadDir(s.completedPath)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", fn, err)
	}

	images := make([]Image, 0, len(entries))
	timeFormat := "02.01.2006 15:04:05"
	for _, e := range entries {
		filename := e.Name()

		fileInfo, err := times.Stat(s.completedPath + filename)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", fn, err)
		}

		updatedAt := fileInfo.ModTime().Format(timeFormat)

		createdAt := "-"
		if fileInfo.HasBirthTime() {
			createdAt = fileInfo.BirthTime().Format(timeFormat)
		}

		images = append(images, Image{
			Name:      filename,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	return images, nil
}

// Search searches image on disk.
func (s *Storage) Search(filename string) (*os.File, error) {
	const fn = "drive.Search"
	file, err := os.Open(s.completedPath + filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return file, nil
}

// FileExists checks file exists.
func (s *Storage) FileExists(filename string) (bool, error) {
	file, err := s.Search(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer file.Close()
	return true, nil
}
