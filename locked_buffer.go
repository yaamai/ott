package main

import (
    "bytes"
    "sync"
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
func (b *LockedBuffer) ReadToPattern(pattern []byte) []byte {
    for {
        // TODO: timeout
        pos := bytes.Index(b.Bytes(), pattern)
        if pos != -1 {
            buffer := make([]byte, pos)
            b.Read(buffer)
            // TODO: handle read err, len
            return buffer
        }
    }
}

