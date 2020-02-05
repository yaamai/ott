package main

import (
	"bytes"
	"errors"
	"sync"
	"time"
)

type LockedBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *LockedBuffer) Bytes() []byte {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Bytes()
}

func (b *LockedBuffer) Reset() {
	b.m.Lock()
	defer b.m.Unlock()
	b.b.Reset()
}

func (b *LockedBuffer) Grow(n int) {
	b.m.Lock()
	defer b.m.Unlock()
	b.b.Grow(n)
}

func (b *LockedBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

func (b *LockedBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

// search pattern and copy-and-read data and return
func (b *LockedBuffer) ReadToPattern(pattern []byte) ([]byte, error) {
	for retry := 0; retry < 100; retry += 1 {
		pos := bytes.Index(b.Bytes(), pattern)
		if pos != -1 {
			buffer := make([]byte, pos)
			l, err := b.Read(buffer)
			if err != nil {
				return nil, err
			}
			if l != pos {
				return nil, errors.New("failed to read from buffer")
			}
			return buffer, nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	return nil, errors.New("timeout waiting pattern present")
}
