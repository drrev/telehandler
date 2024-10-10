// Adapted from sync/cond_test.go

package safe

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCondBroadcast(t *testing.T) {
	c := NewCond()
	n := 400
	running := make(chan int, n)
	awake := make(chan int, n)

	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithTimeout(ctx, time.Until(deadline))
	}
	t.Cleanup(cancel)
	exit := false
	for i := range n {
		go func(g int) {
			seq := int64(0)
			for !exit {
				running <- g
				seq = c.Wait(ctx, seq)
				awake <- g
			}
		}(i)
	}
	for i := range n {
		for range n {
			<-running // Will deadlock unless n are running.
		}
		if i == n-1 {
			exit = true
		}
		select {
		case <-awake:
			t.Fatal("goroutine not asleep")
		default:
		}
		c.Broadcast()
		seen := make([]bool, n)
		for range n {
			g := <-awake
			if seen[g] {
				t.Fatal("goroutine woke up twice")
			}
			seen[g] = true
		}
	}
	select {
	case <-running:
		t.Fatal("goroutine did not exit")
	default:
	}
	c.Broadcast()
}

func BenchmarkCond1(b *testing.B) {
	benchmarkCond(b, 1)
}

func BenchmarkCond2(b *testing.B) {
	benchmarkCond(b, 2)
}

func BenchmarkCond4(b *testing.B) {
	benchmarkCond(b, 4)
}

func BenchmarkCond8(b *testing.B) {
	benchmarkCond(b, 8)
}

func BenchmarkCond16(b *testing.B) {
	benchmarkCond(b, 16)
}

func BenchmarkCond32(b *testing.B) {
	benchmarkCond(b, 32)
}

func benchmarkCond(b *testing.B, waiters int) {
	b.Helper()
	c := NewCond()
	done := make(chan bool, waiters)

	idVal := 0
	mu := sync.Mutex{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		select {
		case <-ctx.Done():
			b.Log("Timeout occurred")
			b.FailNow()
		default:
			cancel()
		}
	}()

	for routine := 0; routine < waiters+1; routine++ {
		go func() {
			seq := int64(0)
			for i := 0; i < b.N; i++ {
				mu.Lock()
				if id := idVal; id == -1 {
					mu.Unlock()
					break
				}
				idVal++
				id := idVal

				if id == waiters+1 {
					idVal = 0
					mu.Unlock()
					c.Broadcast()
				} else {
					mu.Unlock()
					seq = c.Wait(ctx, seq)
				}
			}
			mu.Lock()
			idVal = -1
			mu.Unlock()
			c.Broadcast()
			done <- true
		}()
	}
	for routine := 0; routine < waiters+1; routine++ {
		<-done
	}
}
