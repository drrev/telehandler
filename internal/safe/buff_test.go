package safe

import (
	"bytes"
	"context"
	"io"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestBuffer(t *testing.T) {
	// the number of listeners
	N := 200

	chacha := rand.NewChaCha8([32]byte{})
	b := &Buffer{
		cond:   NewCond(),
		buf:    bytes.NewBuffer(make([]byte, 0, 4096)),
		m:      &sync.RWMutex{},
		closed: false,
	}

	buff := make([]byte, 4096)
	chacha.Read(buff)

	readAndCompare := func(wantErr error) {
		data := make([]byte, 4096)
		if got, err := b.ReadAt(data, 0); got != 4096 {
			t.Errorf("Buffer.ReadAt() = %v, want %v", got, 4096)
		} else if wantErr != err {
			t.Errorf("Buffer.ReadAt() error = %v, wantErr %v", err, wantErr)
		}

		if slices.Compare(buff, data) != 0 {
			t.Errorf("Buffer.ReadAt() data = %#v, want %#v", data, buff)
		}
	}

	wg := &sync.WaitGroup{}

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			// Buffer.Wait(), Buffer.ReadAt(), Buffer.Write()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			defer wg.Done()
			b.Wait(ctx, 0)
			readAndCompare(nil)
		}(wg)
	}

	// Buffer.Write()
	if got, err := b.Write(buff); got != 4096 {
		t.Errorf("Buffer.Write() = %v, want %v", got, 4096)
	} else if err != nil {
		t.Errorf("Buffer.Write() error = %v, wantErr %v", err, false)
	}

	// Buffer.Len()
	if got := b.Len(); got != 4096 {
		t.Errorf("Buffer.Len() = %v, want %v", got, 4096)
	}

	wg.Wait()

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			// Buffer.Wait(), Buffer.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			defer wg.Done()
			b.Wait(ctx, 1)
			readAndCompare(io.EOF)
		}(wg)
	}

	b.Close()
	wg.Wait()
}
