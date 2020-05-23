package main

import (
	"bytes"
	"errors"
	"go.uber.org/zap"
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

func (b *LockedBuffer) Len() (n int) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Len()
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
		zap.S().Debug("ReadToPattern", retry, pos, b.Bytes(), pattern)
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

		time.Sleep(100 * time.Millisecond)
	}

	return nil, errors.New("timeout waiting pattern present")
}

func (b *LockedBuffer) ReadBetweenPattern(startPattern, endPattern []byte) ([]byte, error) {

	buf := b.Bytes()
	startPos := bytes.Index(buf, startPattern)
	if startPos == -1 {
		return nil, nil
	}
	startPos += len(startPattern)
	endPos := bytes.Index(buf[startPos:], endPattern)
	if endPos == -1 {
		return nil, nil
	}
	endPos += startPos

	// skip startPattern
	ignoreBuf := make([]byte, startPos)
	l, err := b.Read(ignoreBuf)
	if err != nil {
		return nil, err
	}

	// read target data
	resultLen := endPos - startPos
	resultBuf := make([]byte, resultLen)
	l, err = b.Read(resultBuf)
	if err != nil {
		return nil, err
	}
	if l != resultLen {
		return nil, errors.New("failed to read from buffer")
	}

	// TODO: read endPattern
	return resultBuf, nil
}

func (b *LockedBuffer) WaitStable(retryMax int, wait time.Duration) (int64, error) {
	oldlen := b.Len()
	count := 0
	waitTime := int64(0)
	for retry := 0; retry < retryMax; retry += 1 {
		l := b.Len()
		if oldlen > 0 && count > 5 {
			return waitTime, nil
		}
		if l == oldlen {
			count += 1
		}
		oldlen = l
		time.Sleep(wait)
		waitTime += wait.Milliseconds()
	}

	return waitTime, errors.New("buffer wait timeout")
}
