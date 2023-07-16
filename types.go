package filecache

import "errors"

var (
	// ErrQueueClosed is returned when enqueue/dequeue is being called
	// after Close/Destroy.
	ErrQueueClosed = errors.New("queue doesn't exist")
	// ErrNilType is returned when the encoder is not enabled/invalid.
	ErrNilType = errors.New("cannot use nil type")
	// ErrExceedsSize represents the error when the file size exceeds the given size.
	ErrExceedsSize = errors.New("exceeds max size")
)

// EncodingType represents the supported encoding types.
type EncodingType int

// Encoding types.
const (
	Gob EncodingType = iota
	JSON
)

// DropMode represensts the supported drop modes.
type DropMode int

// Drop modes.
const (
	HeadDrop DropMode = iota
	TailDrop
)
