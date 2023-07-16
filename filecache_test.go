//go:build !windows

// nolint
package filecache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_FileEnqueueDequeueSync(t *testing.T) {

	name, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	dbPath = name
	var result int

	fileAPI, err := NewFileCache("test1", OptionMaxElements(1000), OptionTTL(20*time.Hour))
	require.Nil(t, err, "filecache should not return an error")
	require.NotNil(t, fileAPI, "filecache should not be nil")

	defer os.RemoveAll(name + "/test1")

	require.Equal(t, 20*time.Hour, fileAPI.duration)

	fileAPI.Run(context.Background())
	time.Sleep(10 * time.Millisecond)

	var strchan chan string
	err = fileAPI.enqueue(&strchan)
	require.NotNil(t, err, "invalid type")

	for i := 0; i < 1000; i++ {
		err = fileAPI.enqueue(i)
		require.Nil(t, err, fmt.Sprintf("enqueue should not return an error: %d", i))
	}

	require.Equal(t, 1000, int(fileAPI.lockedLength()), "length should equal 1000")

	err = fileAPI.Close()
	require.Nil(t, err, "close should not return an error")

	fileAPI, err = NewFileCache("test1", OptionMaxElements(1000))
	require.Nil(t, err, "filecache should not return an error")
	require.NotNil(t, fileAPI, "filecache should not be nil")

	require.Equal(t, 24*time.Hour, fileAPI.duration)

	require.Equal(t, 1000, int(fileAPI.lockedLength()), "length should equal 1000")

	for i := 0; i < 1000; i++ {
		err = fileAPI.Dequeue(&result)
		require.Nil(t, err, fmt.Sprintf("dequeue should not return an error: %d", i))
		require.Equal(t, i, result, fmt.Sprintf("data didn't match, enqueue: %d, dequeue: %d", i, result))
	}

	require.Zero(t, int(fileAPI.lockedLength()), "length should equal 0")

	for i := 0; i < 1010; i++ {
		err = fileAPI.enqueue(i)
		require.Nil(t, err, fmt.Sprintf("enqueue should not return an error: %d", i))
	}

	require.Equal(t, 1000, int(fileAPI.lockedLength()), "length should equal 1000")

	err = fileAPI.enqueue(&strchan)
	require.NotNil(t, err, "invalid type")

	err = fileAPI.Destroy()
	require.Nil(t, err, "destroy should not return an error")

	err = fileAPI.enqueue(2)
	require.NotNil(t, err, "queue should be closed")

	err = fileAPI.Dequeue(2)
	require.NotNil(t, err, "queue should be closed")
}

func Test_FileEnqueueDequeueDropMode(t *testing.T) {

	name, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	dbPath = name

	fileAPI, err := NewFileCache("test23", OptionMaxElements(5), OptionDropMode(TailDrop))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test23")

	err = fileAPI.enqueue(1)
	require.Nil(t, err)
	err = fileAPI.enqueue(2)
	require.Nil(t, err)
	err = fileAPI.enqueue(3)
	require.Nil(t, err)
	err = fileAPI.enqueue(4)
	require.Nil(t, err)
	err = fileAPI.enqueue(5)
	require.Nil(t, err)
	err = fileAPI.enqueue(6)
	require.Nil(t, err)
	err = fileAPI.enqueue(7)
	require.Nil(t, err)

	require.Equal(t, 5, int(fileAPI.lockedLength()))

	var result int
	err = fileAPI.Dequeue(&result)
	require.Nil(t, err)
	require.Equal(t, 1, result)

	err = fileAPI.Destroy()
	require.Nil(t, err)

	fileAPI, err = NewFileCache("test34", OptionMaxElements(5), OptionDropMode(HeadDrop))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test34")

	err = fileAPI.enqueue(1)
	require.Nil(t, err)
	err = fileAPI.enqueue(2)
	require.Nil(t, err)
	err = fileAPI.enqueue(3)
	require.Nil(t, err)
	err = fileAPI.enqueue(4)
	require.Nil(t, err)
	err = fileAPI.enqueue(5)
	require.Nil(t, err)
	err = fileAPI.enqueue(6)
	require.Nil(t, err)
	err = fileAPI.enqueue(7)
	require.Nil(t, err)

	require.Equal(t, 5, int(fileAPI.lockedLength()))

	var result1 int
	err = fileAPI.Dequeue(&result1)
	require.Nil(t, err)
	require.Equal(t, 3, result1)

	err = fileAPI.Destroy()
	require.Nil(t, err)

	fileAPI, err = NewFileCache("test35", OptionMaxElements(1), OptionDropMode(45))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test35")

	err = fileAPI.enqueue(1)
	require.Nil(t, err)
	err = fileAPI.enqueue(2)
	require.NotNil(t, err)

	err = fileAPI.Destroy()
	require.Nil(t, err)
}

func Test_FileEnqueueDequeueEncoding(t *testing.T) {

	name, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	dbPath = name

	fileAPI, err := NewFileCache("test1101", OptionEncodingType(JSON))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test1101")

	err = fileAPI.enqueue(1)
	require.Nil(t, err)

	require.Equal(t, 1, int(fileAPI.lockedLength()))

	var result int
	err = fileAPI.Dequeue(&result)
	require.Nil(t, err)
	require.Equal(t, 1, result)

	err = fileAPI.Destroy()
	require.Nil(t, err)

	fileAPI, err = NewFileCache("test1102", OptionEncodingType(Gob))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test1102")

	err = fileAPI.enqueue(1)
	require.Nil(t, err)

	require.Equal(t, 1, int(fileAPI.lockedLength()))

	var result1 int
	err = fileAPI.Dequeue(&result1)
	require.Nil(t, err)
	require.Equal(t, 1, result1)

	err = fileAPI.Destroy()
	require.Nil(t, err)

	fileAPI, err = NewFileCache("test1103", OptionEncodingType(40))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test1103")

	err = fileAPI.enqueue(1)
	require.NotNil(t, err)

	err = fileAPI.Destroy()
	require.Nil(t, err)

	fileAPI, err = NewFileCache("test1104", OptionEncodingType(Gob))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test1104")

	err = fileAPI.enqueue(1)
	require.Nil(t, err)

	require.Equal(t, 1, int(fileAPI.lockedLength()))

	fileAPI.encodingType = 20

	var result2 int
	err = fileAPI.Dequeue(&result2)
	require.NotNil(t, err)

	err = fileAPI.Destroy()
	require.Nil(t, err)
}

func Test_FileEnqueueDequeueMaxSize(t *testing.T) {

	name, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	dbPath = name

	fileAPI, err := NewFileCache("test1201", OptionMaxSize(4))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test1201")

	err = fileAPI.enqueue(1)
	require.NotNil(t, err)
	require.Equal(t, ErrExceedsSize, err)

	require.Equal(t, 0, int(fileAPI.lockedLength()))

	err = fileAPI.Destroy()
	require.Nil(t, err)

	fileAPI, err = NewFileCache("test1202", OptionMaxSize(1000))
	require.Nil(t, err)
	require.NotNil(t, fileAPI)

	defer os.RemoveAll(name + "/test1202")

	err = fileAPI.enqueue(1)
	require.Nil(t, err)

	require.Equal(t, 1, int(fileAPI.lockedLength()))

	var result1 int
	err = fileAPI.Dequeue(&result1)
	require.Nil(t, err)
	require.Equal(t, 1, result1)

	err = fileAPI.Destroy()
	require.Nil(t, err)
}

func Test_FileEnqueueDequeueAsync(t *testing.T) {

	name, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	dbPath = name
	var result int

	fileAPI, err := NewFileCache("test3")
	require.Nil(t, err, "filecache should not return an error")
	require.NotNil(t, fileAPI, "filecache should not be nil")

	fileAPI.Run(context.Background())
	time.Sleep(10 * time.Millisecond)

	fileAPI.Enqueue(1)
	fileAPI.Enqueue(2)

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 2, int(fileAPI.lockedLength()), "length should equal 2")

	err = fileAPI.Dequeue(&result)
	require.Nil(t, err, fmt.Sprintf("dequeue should not return an error"))
	require.Equal(t, 1, result, fmt.Sprintf("data didn't match, enqueue: 1, dequeue: %d", result))

	err = fileAPI.Dequeue(&result)
	require.Nil(t, err, fmt.Sprintf("dequeue should not return an error"))
	require.Equal(t, 2, result, fmt.Sprintf("data didn't match, enqueue: 2, dequeue: %d", result))

	err = fileAPI.Destroy()
	require.Nil(t, err, "destroy should not return an error")

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		err := <-fileAPI.Errors()
		require.NotNil(t, err)
		require.Equal(t, ErrQueueClosed, err)
		wg.Done()
	}(&wg)

	time.Sleep(10 * time.Millisecond)

	fileAPI.Enqueue(1)
	wg.Wait()
}

func Test_FileEnqueueBlock(t *testing.T) {

	name, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	dbPath = name

	objectChannelSize = 2
	defer func() {
		objectChannelSize = 100
	}()

	fileAPI, err := NewFileCache("test4", OptionMaxElements(1000))
	require.Nil(t, err, "filecache should not return an error")
	require.NotNil(t, fileAPI, "filecache should not be nil")

	var wg sync.WaitGroup

	wg.Add(3)
	var errCount int32
	go func(wg *sync.WaitGroup) {
		for {
			err := <-fileAPI.Errors()
			if err != nil {
				atomic.AddInt32(&errCount, 1)
			}
			wg.Done()
		}
	}(&wg)

	time.Sleep(10 * time.Millisecond)

	fileAPI.Enqueue(1)
	fileAPI.Enqueue(2)
	fileAPI.Enqueue(3)
	fileAPI.Enqueue(4)
	fileAPI.Enqueue(5)

	wg.Wait()

	e := atomic.SwapInt32(&errCount, 0)
	if e != 3 {
		t.Fatalf("unexpected errcount, expected: 3, actual: %v", e)
	}

	require.Equal(t, len(fileAPI.objectCh), 2, "channel should be buffered and non-blocking")

	err = fileAPI.Destroy()
	require.Nil(t, err, "destroy should not return an error")
}

func Test_validateSize(t *testing.T) {

	path, err := os.MkdirTemp("", "")
	require.Nil(t, err)

	defer os.RemoveAll(path)

	err = os.WriteFile(filepath.Join(path, "test666.txt"), make([]byte, 200), 0666)
	require.Nil(t, err)

	err = os.WriteFile(filepath.Join(path, "test777.txt"), make([]byte, 200), 0666)
	require.Nil(t, err)

	size, err := dbSize(path)
	require.Nil(t, err)
	require.Equal(t, int64(400), size)

	err = validateSize(path, 600, make([]byte, 150))
	require.Nil(t, err)

	err = validateSize(path, 600, make([]byte, 200))
	require.Nil(t, err)

	err = validateSize(path, 600, make([]byte, 201))
	require.NotNil(t, err)
	require.Equal(t, ErrExceedsSize, err)

	err = validateSize(path, 0, make([]byte, 401))
	require.Nil(t, err)
}
