package filecache

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/beeker1121/goque"
	"github.com/sibicramesh/limitedbuf-go"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/zap"
)

var _ Queue = &FileCache{}

var (
	dbPath            string
	objectChannelSize = 10000
	errChannelSize    = 16
)

// In most cases we don't need to modify these options as
// these are relevant when we want to read from the DB. The
// filecache implementation is intended to be write heavy and
// implements the queue mechanishm where read is supposed to
// happen occasionally. These options reduces the memory
// footprint significantly at the expense of read complexity.
// Ref# https://github.com/google/leveldb/tree/master/doc
// Ref# https://pkg.go.dev/github.com/syndtr/goleveldb@v1.0.0/leveldb/opt
var defaultOptions = &opt.Options{
	WriteBuffer:       100 * opt.KiB,
	BlockCacher:       opt.NoCacher,
	DisableBlockCache: true,
}

// ConfigureDBPath updates the db path
func ConfigureDBPath(path string) {
	dbPath = path
}

// FileCache implements Queue interface.
type FileCache struct {
	name              string
	t                 reflect.Type
	maxElements       uint64
	encodedBufferSize int
	duration          time.Duration
	que               *goque.PrefixQueue
	objectCh          chan interface{}
	encodingType      EncodingType
	path              string
	maxSize           int64
	dropMode          DropMode
	errChan           chan error

	queueLock sync.RWMutex
	sync.Mutex
}

// NewFileCache creates a new filecache handler to handle arbitrary types to be stored in on disk.
// It uses goque which is backend by go's port of levelDB.
// It implements Queue interface and has tools to encode/decode using msgpack.
// This package also runs a cleaner job which every 24Hrs, will remove the existing DB completely
// and create a new DB instance.
// name - used as the filename in the db for this type.
// All methods are thread-safe.
func NewFileCache(name string, options ...Option) (*FileCache, error) {

	fc := &FileCache{
		name:     name,
		objectCh: make(chan interface{}, objectChannelSize),
		path:     filepath.Join(dbPath, name),
		errChan:  make(chan error, errChannelSize),
	}

	for _, opt := range options {
		opt(fc)
	}

	if fc.duration == 0 {
		fc.duration = 24 * time.Hour
	}

	var err error
	fc.que, err = openOrRecoverQueue(fc.path)
	if err != nil {
		return nil, err
	}

	return fc, nil
}

// Name returns name used by the caller to identify the filecache.
func (f *FileCache) Name() string {
	return f.name
}

// Run runs cleaner job and starts listener for new entries.
func (f *FileCache) Run(ctx context.Context) error {

	go f.cleaner(ctx)

	go func() {
		for {
			select {
			case o := <-f.objectCh:
				if err := f.enqueue(o); err != nil {
					f.sendError(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Enqueue inserts a new entry to the queue. Non-blocking.
func (f *FileCache) Enqueue(obj interface{}) error {

	select {
	case f.objectCh <- obj:
	default:
		f.sendError(fmt.Errorf("queue buffer is full, object dropped: %v", obj))
	}

	return nil
}

// Dequeue deletes removes the first entry in the queue.
func (f *FileCache) Dequeue(obj interface{}) error {

	item, err := f.lockedDequeue(obj)
	if err != nil {
		return err
	}

	return unmarshal(f.encodingType, item.Value, obj)
}

// Errors returns the error pipe.
func (f *FileCache) Errors() chan error {
	return f.errChan
}

func (f *FileCache) sendError(err error) {

	select {
	case f.errChan <- err:
	default:
	}
}

// Encode dequeues all the entries and encodes them with the reflect.Type.
func (f *FileCache) Encode() ([]byte, error) {

	if f.t == nil {
		return nil, ErrNilType
	}

	var encodedData limitedbuf.Buffer
	if f.encodedBufferSize > 0 {
		encodedData = limitedbuf.NewBuffer(make([]byte, 0), f.encodedBufferSize)
	} else {
		encodedData = bytes.NewBuffer(nil)
	}

	enc := msgpack.NewEncoder(encodedData)

	for {
		vPtr := reflect.New(f.t)

		err := f.Dequeue(vPtr.Interface())
		switch err {
		case nil:
		case io.EOF, goque.ErrEmpty, goque.ErrOutOfBounds:
			return encodedData.Bytes(), nil
		default:
			return nil, err
		}

		// We encode the pointer to type.
		if err := enc.Encode(vPtr.Interface()); err != nil {
			if err == limitedbuf.ErrWriteExceedsBufCap {
				return encodedData.Bytes(), nil
			}
			return nil, err
		}
	}
}

// EncodeToReader functions like Encode but returns a reader.
func (f *FileCache) EncodeToReader() (io.Reader, int64, error) {
	// We expect that encoded types will not be huge and that we can encode them
	// into a byte[] that will fit in memory.
	buf, err := f.Encode()
	if err != nil {
		return nil, 0, err
	}

	return bytes.NewReader(buf), int64(len(buf)), nil
}

// Decode decodes the given data and returns the list as []interface.
// Caller has to type assert the interface to the respected types.
func (f *FileCache) Decode(data []byte) ([]interface{}, error) {

	if f.t == nil {
		return nil, ErrNilType
	}

	var result []interface{}

	reader := bytes.NewReader(data)
	dec := msgpack.NewDecoder(reader)

	for {
		vPtr := reflect.New(f.t)

		err := dec.Decode(vPtr.Interface())
		switch err {
		case nil:
		case io.EOF:
			return result, nil
		default:
			return nil, err
		}

		// This is fine since we encoded pointer to type.
		result = append(result, vPtr.Elem().Interface())
	}
}

// Close closes the db.
func (f *FileCache) Close() error {

	f.Lock()
	defer f.Unlock()

	if err := f.queue().Close(); err != nil {
		return err
	}

	f.setQueue(nil)
	return nil
}

// Destroy drops the db, removes any leftover files/directories and frees the queue.
func (f *FileCache) Destroy() error {

	f.Lock()
	defer f.Unlock()

	if err := f.queue().Drop(); err != nil {
		return err
	}

	f.setQueue(nil)
	return nil
}

func (f *FileCache) enqueue(obj interface{}) error {

	if f.maxElements > 0 && f.lockedLength() == f.maxElements {

		switch f.dropMode {
		case HeadDrop:
			if _, err := f.lockedDequeue(obj); err != nil {
				return err
			}
		case TailDrop:
			return nil
		default:
			return fmt.Errorf("unknown drop mode: %v", f.dropMode)
		}
	}

	if _, err := f.lockedEnqueue(obj); err != nil {
		return err
	}

	return nil
}

func (f *FileCache) cleaner(ctx context.Context) {

	for {
		select {
		case <-time.After(f.duration):

			if err := f.Destroy(); err != nil {
				zap.L().Error("unable to drop queue", zap.Error(err))
				return
			}

			q, err := goque.OpenPrefixQueue(f.path, defaultOptions)
			if err != nil {
				zap.L().Error("unable to create queue", zap.Error(err))
				return
			}

			f.setQueue(q)

		case <-ctx.Done():
			return
		}
	}
}

func (f *FileCache) lockedEnqueue(obj interface{}) (*goque.Item, error) {

	f.Lock()
	defer f.Unlock()

	if f.queue() == nil {
		return nil, ErrQueueClosed
	}

	prefix, err := extractName(obj)
	if err != nil {
		return nil, err
	}

	value, err := marshal(f.encodingType, obj)
	if err != nil {
		return nil, err
	}

	if err := validateSize(f.path, f.maxSize, value); err != nil {
		return nil, err
	}

	return f.queue().Enqueue(prefix, value)
}

func (f *FileCache) lockedDequeue(obj interface{}) (*goque.Item, error) {

	f.Lock()
	defer f.Unlock()

	if f.queue() == nil {
		return nil, ErrQueueClosed
	}

	prefix, err := extractName(obj)
	if err != nil {
		return nil, err
	}

	return f.queue().Dequeue(prefix)
}

func (f *FileCache) lockedLength() uint64 {

	f.Lock()
	defer f.Unlock()

	if f.queue() == nil {
		return 0
	}

	return f.queue().Length()
}

func (f *FileCache) queue() *goque.PrefixQueue {

	f.queueLock.RLock()
	defer f.queueLock.RUnlock()

	return f.que
}

func (f *FileCache) setQueue(queue *goque.PrefixQueue) {

	f.queueLock.Lock()
	f.que = queue
	f.queueLock.Unlock()
}

// Use logic similar to goque.checkGoqueType to check for incompatible old-format db.
// Unfortunately, if OpenPrefixQueue fails, then goque will not close its file lock
// (on Windows, at least), and os.RemoveAll will not even work after that.
// Issue filed here https://github.com/beeker1121/goque/issues/28
func checkAndDeleteIncompatibleQueue(path string) {
	if f, err := os.OpenFile(filepath.Join(path, "GOQUE"), os.O_RDONLY, 0); err == nil {
		const goquePrefixQueue = 3
		fb := make([]byte, 1)
		if _, err = f.Read(fb); err == nil {
			if fb[0] != goquePrefixQueue {
				// remove old incompatible db
				zap.L().Warn("removing existing incompatible queue", zap.String("path", path))
				f.Close() //nolint
				if err = os.RemoveAll(path); err != nil {
					zap.L().Warn("failed to remove incompatible queue", zap.String("path", path), zap.Error(err))
				}
				return
			}
		}
		f.Close() //nolint
	}
}

func openOrRecoverQueue(path string) (*goque.PrefixQueue, error) {

	checkAndDeleteIncompatibleQueue(path)

	q, err := goque.OpenPrefixQueue(path, defaultOptions)
	if err != nil {
		zap.L().Error("unable to open queue, trying to recover",
			zap.String("path", path),
			zap.Error(err))

		db, err := leveldb.RecoverFile(path, nil)
		if err != nil {
			zap.L().Error("unable to recover db, recreating",
				zap.String("path", path),
				zap.Error(err))

			if err := os.RemoveAll(path); err != nil {
				return nil, err
			}
		} else {
			zap.L().Info("Queue recovered", zap.String("path", path))
			db.Close() // nolint: errcheck
		}

		return goque.OpenPrefixQueue(path, defaultOptions)
	}

	return q, nil
}

func extractName(obj interface{}) ([]byte, error) {

	if obj == nil {
		return nil, errors.New("object is nil")
	}

	typ := reflect.TypeOf(obj)

	for typ.Kind() == reflect.Ptr ||
		typ.Kind() == reflect.Interface {

		typ = typ.Elem()
	}

	name := typ.Name()
	if name == "" {
		return nil, errors.New("type has no name")
	}

	return []byte(name), nil
}

func marshal(e EncodingType, obj interface{}) ([]byte, error) {

	switch e {
	case JSON:
		return json.Marshal(obj)

	case Gob:
		var buffer bytes.Buffer
		enc := gob.NewEncoder(&buffer)
		if err := enc.Encode(obj); err != nil {
			return nil, err
		}

		return buffer.Bytes(), nil

	default:
		return nil, fmt.Errorf("unknown encoding type: %v", e)
	}
}

func unmarshal(e EncodingType, data []byte, dest interface{}) error {

	switch e {
	case JSON:
		return json.Unmarshal(data, dest)

	case Gob:
		buffer := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buffer)

		return dec.Decode(dest)

	default:
		return fmt.Errorf("unknown decoding type: %v", e)
	}
}

func validateSize(path string, maxsz int64, data []byte) error {

	if maxsz <= 0 {
		return nil
	}

	currsz, err := dbSize(path)
	if err != nil {
		return err
	}

	if int64(len(data))+currsz > maxsz {
		return ErrExceedsSize
	}

	return nil
}

func dbSize(path string) (int64, error) {

	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return err
	})

	return size, err
}
