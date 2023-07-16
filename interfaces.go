package filecache

import (
	"context"
	"errors"
	"io"
)

// ErrTooLarge can be returned from `Encode` to indicate that data is large and EncodeToReader should be used
var ErrTooLarge = errors.New("Data too large")

// Queue interface has all necessary methods to enqueue, dequeue,
// encode, decode entries.
type Queue interface {
	Name() string
	Run(ctx context.Context) error
	Enqueue(element interface{}) error
	Dequeue(element interface{}) error
	Close() error
	Destroy() error
	Errors() chan error

	Encode() ([]byte, error)
	Decode(data []byte) ([]interface{}, error)

	EncodeToReader() (io.Reader, int64, error)
}
