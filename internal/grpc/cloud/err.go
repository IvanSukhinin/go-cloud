package cloud

import (
	"errors"
	"fmt"
)

var (
	ErrInternal      = errors.New("internal error")
	ErrNotExist      = errors.New("image doesn't exist")
	ErrEmptyFilename = errors.New("filename is empty")
)

type ErrImageExt struct {
	ext map[string]struct{}
}

func (e *ErrImageExt) Error() string {
	return fmt.Sprintf("unsupported ext. available exts: %v", e.ext)
}

type ErrImageMaxSize struct {
	maxImageSize int
}

func (e *ErrImageMaxSize) Error() string {
	return fmt.Sprintf("image is too large. max size: %d", e.maxImageSize)
}
