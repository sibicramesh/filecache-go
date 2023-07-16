package filecache

import (
	"reflect"
	"time"
)

// An Option represents a Filecache option.
type Option func(*FileCache)

// OptionEncoderConfig used to set the type and the buffer size.
// obj - the object type to use with Encode/Decode.
// bufferSize - when Encode is called, we will limit the encoded data size with the
// provided bytes size. [0 - infinite entries]
// Default: Disabled.
func OptionEncoderConfig(obj interface{}, bufferSize int) Option {
	return func(f *FileCache) {
		f.t = reflect.TypeOf(obj)
		f.encodedBufferSize = bufferSize
	}
}

// OptionEncodingType used to set encoding type. Supported options
// are JSON and Gob.
// Default: Gob.
func OptionEncodingType(e EncodingType) Option {
	return func(f *FileCache) {
		f.encodingType = e
	}
}

// OptionMaxElements used to set the max elements.
// maxElements - the db will be restricted to not exceed this limit and any enqueues after this
// will call dequeue and enqueues the new entry. It is validated before OptionMaxSize.
// Default: Infinite.
func OptionMaxElements(maxElements uint64) Option {
	return func(f *FileCache) {
		f.maxElements = maxElements
	}
}

// OptionMaxSize used to set the max data size on disk.
// maxSize - the db will be restricted to not exceed this limit and any enqueues after this
// will be dropped with ErrExceedsSize error.
// In most cases, you need to use OptionMaxElements over this.
// Default: Disabled.
func OptionMaxSize(maxSize int64) Option {
	return func(f *FileCache) {
		f.maxSize = maxSize
	}
}

// OptionDropMode used to set the drop mode when maxElements limit has reached.
// Default: HeadDrop.
func OptionDropMode(mode DropMode) Option {
	return func(f *FileCache) {
		f.dropMode = mode
	}
}

// OptionTTL used to set the TTL of this DB.
// Default: 24h.
func OptionTTL(duration time.Duration) Option {
	return func(f *FileCache) {
		f.duration = duration
	}
}
