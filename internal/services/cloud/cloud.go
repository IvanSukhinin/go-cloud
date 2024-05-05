package cloud

import (
	"bytes"
	"cloud/internal/storage/drive"
	"fmt"
	"log/slog"
	"os"
)

type Cloud struct {
	log     *slog.Logger
	storage Storage
}

func New(
	log *slog.Logger,
	drive Storage,
) *Cloud {
	return &Cloud{
		log:     log,
		storage: drive,
	}
}

type Storage interface {
	Save(filename string, buf bytes.Buffer) error
	List() ([]drive.Image, error)
	Search(filename string) (*os.File, error)
	FileExists(filename string) (bool, error)
}

func (c *Cloud) Upload(filename string, buf bytes.Buffer) error {
	const fn = "services.cloud.Upload"

	// some business logic

	err := c.storage.Save(filename, buf)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil
}

func (c *Cloud) CanUpload(filename string) (bool, error) {
	const fn = "services.cloud.CanUpload"
	isExist, err := c.storage.FileExists(filename)
	if err != nil {
		return false, fmt.Errorf("%s: %w", fn, err)
	}
	return !isExist, err
}

func (c *Cloud) List() ([]drive.Image, error) {
	const fn = "services.cloud.List"

	// some business logic

	images, err := c.storage.List()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return images, nil
}

func (c *Cloud) Search(filename string) (*os.File, error) {
	const fn = "services.cloud.Search"

	// some business logic

	file, err := c.storage.Search(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return file, nil
}
