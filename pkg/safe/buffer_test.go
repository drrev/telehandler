package safe

import (
	"fmt"
	"io"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestWriteAndNotify(t *testing.T) {
	nb := NewNotifyingBuffer()
	data := []byte("Hello, World!")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		// Check if notify channel is closed (which means buffer has changed)
		ch := nb.Wait()
		<-ch
		wg.Done()
	}()

	n, err := nb.Write(data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, but wrote %d", len(data), n)
	}
	nb.Close()

	wg.Wait()
}

func TestCloseAndWait(t *testing.T) {
	nb := NewNotifyingBuffer()
	data := []byte("Hello, World!")
	_, err := nb.Write(data)
	if err != nil {
		t.Errorf("NotifyingBuffer.Write() unexpected error: %v", err)
	}
	err = nb.Close()
	if err != nil {
		t.Errorf("NotifyingBuffer.Close() unexpected error: %v", err)
	}
	err = nb.Close()
	if err != nil {
		t.Errorf("NotifyingBuffer.Close() unexpected error: %v", err)
	}
	// Check if notify channel is closed (which means the writer is closed)
	<-nb.Wait()
}

func TestReader(t *testing.T) {
	nb := NewNotifyingBuffer()
	data := []byte("Hello, World!")
	_, err := nb.Write(data)
	if err != nil {
		t.Errorf("NotifyingBuffer.Write() unexpected error: %v", err)
	}

	r := nb.Reader()
	buf := []byte{}

	n, err := r.Read(buf)
	if err != nil {
		t.Errorf("NotifyingBufferReader.Read() expected no error, got %v", err)
	}
	if n != 0 {
		t.Error("NotifyingBufferReader.Read() expected no data to be read")
	}

	buf = make([]byte, len(data))
	n, err = r.Read(buf)
	if err != nil {
		t.Errorf("NotifyingBufferReader.Read() expected no error, got %v", err)
	}
	if n != len(data) {
		t.Errorf("NotifyingBufferReader.Read() expected %d data to be read", len(data))
	}

	if !slices.Equal(data, buf) {
		t.Errorf("NotifyingBufferReader.Read() expected %s, got %s", data, buf)
	}

	r.Close()
	n, err = r.Read(buf)
	if err != io.EOF {
		t.Errorf("NotifyingBufferReader.Read() expected EOF, got %v", err)
	}
	if n != 0 {
		t.Error("NotifyingBufferReader.Read() expected no data to be read")
	}
	r.Close()

	r = nb.Reader()
	nb.Close()

	_, _ = r.Read(buf)
	n, err = r.Read(buf)
	if err != io.EOF {
		t.Errorf("NotifyingBufferReader.Read() expected EOF, got %v", err)
	}
	if n != 0 {
		t.Error("NotifyingBufferReader.Read() expected no data to be read")
	}

	r.Close()
}

func TestConcurrentWriteAndRead(t *testing.T) {
	nb := NewNotifyingBuffer()
	data1 := []byte("Hello, ")
	data2 := []byte("World!")
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = nb.Write(data1)
	}()
	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		_, _ = nb.Write(data2)
	}()
	wg.Wait()

	r := nb.Reader()
	buf := make([]byte, len(data1)+len(data2))
	n, err := r.Read(buf)
	if err != nil {
		t.Errorf("NotifyingBufferReader.Read(): %v", err)
	}
	if n != len(data1)+len(data2) {
		t.Errorf("NotifyingBufferReader.Read() expected to read %d bytes, but read %d", len(data1)+len(data2), n)
	}

	if data := buf[:len(data1)]; !slices.Equal(data1, data) {
		t.Errorf("NotifyingBufferReader.Read() expected %s, got %s", data1, data)
	}

	if data := buf[len(data1):]; !slices.Equal(data2, data) {
		t.Errorf("NotifyingBufferReader.Read() expected %s, got %s", data1, data)
	}
}

func TestCloseWhileReading(t *testing.T) {
	nb := NewNotifyingBuffer()
	r := nb.Reader()
	errCh := make(chan error)

	go func() {
		_, err := r.Read(nil)
		errCh <- err
	}()

	err := nb.Close()
	if err != nil {
		t.Errorf("NotifyingBuffer.Close() unexpected error: %v", err)
	}

	err = <-errCh
	if err != io.EOF {
		t.Errorf("NotifyingBufferReader.Read() expected EOF but got %v", err)
	}
}

func TestMultipleReaders(t *testing.T) {
	nb := NewNotifyingBuffer()
	data := []byte("Hello, World!")

	_, err := nb.Write(data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	r1 := nb.Reader()
	r2 := nb.Reader()

	buf1 := make([]byte, len(data))
	buf2 := make([]byte, len(data))
	_, err = r1.Read(buf1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, err = r2.Read(buf2)
	if err != nil {
		t.Errorf("NotifyingBufferReader.Read() unexpected error: %v", err)
	}

	if !slices.Equal(data, buf1) {
		t.Errorf("NotifyingBufferReader.Read() reader 1 expected %s, got %s", data, buf1)
	}

	if !slices.Equal(data, buf2) {
		t.Errorf("NotifyingBufferReader.Read() reader 2 expected %s, got %s", data, buf2)
	}
}

func BenchmarkNotifyingBuffer(b *testing.B) {
	rando := makeRandomSlice(b, 65535)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := rando()

		nb := &NotifyingBuffer{
			closed: false,
			buff:   []byte{},
			notify: make(chan struct{}),
			mu:     sync.RWMutex{},
		}

		var eg errgroup.Group
		for i := range 1000 {
			i := i
			eg.Go(func() error {
				read := make([]byte, 65535)
				rd := nb.Reader()
				if _, err := io.ReadFull(rd, read); err != nil {
					return fmt.Errorf("NotifyingBufferReader.Read() worker %d recieved error: %w", i, err)
				}

				for i, v := range read {
					if v != data[i] {
						b.Errorf("NotifyingBufferReader.Read() mismatching slices on worker %d", i)
					}
				}

				return nil
			})
		}

		b.StartTimer()
		offs := 0
		for offs < len(data) {
			n, err := nb.Write(data[offs:])
			if err != nil {
				b.Fatal(err)
			}
			offs += n
		}
		nb.Close()
		if err := eg.Wait(); err != nil {
			b.Error(err)
		}
	}
	b.ReportAllocs()
}

func makeRandomSlice(t interface{ Helper() }, length int) func() []byte {
	t.Helper()
	seed := [32]byte{}
	chacha := rand.NewChaCha8(seed)
	return func() []byte {
		data := make([]byte, length)
		chacha.Read(data)
		return data
	}
}
