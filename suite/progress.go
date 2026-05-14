package suite

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type progressTracker struct {
	bytes atomic.Int64
	total int64
	start time.Time
}

func newProgressTracker(total int64) *progressTracker {
	return &progressTracker{
		total: total,
		start: time.Now(),
	}
}

// run starts a background goroutine that prints aggregate progress to stderr
// every second. Call the returned stop function after all transfers finish.
func (p *progressTracker) run(ctx context.Context, op string) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	var once sync.Once
	p.printLine(op) // show 0% immediately so the line is visible from the start
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.printLine(op)
			case <-stop:
				p.printLine(op) // final update at 100%
				fmt.Fprint(os.Stderr, "\n")
				return
			case <-ctx.Done():
				fmt.Fprint(os.Stderr, "\n")
				return
			}
		}
	}()
	return func() {
		once.Do(func() { close(stop) })
		<-done
	}
}

func (p *progressTracker) printLine(op string) {
	done := p.bytes.Load()
	elapsed := time.Since(p.start).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(done) / 1_000_000 / elapsed
	}
	pct := 0.0
	if p.total > 0 {
		pct = float64(done) / float64(p.total) * 100
	}
	fmt.Fprintf(os.Stderr, "\r  [%-8s] %6.1f / %.1f MB  (%5.1f%%)  @ %5.1f MB/s",
		op,
		float64(done)/1_000_000,
		float64(p.total)/1_000_000,
		pct,
		speed,
	)
}

// wrapReader returns a reader that counts bytes as they pass through.
func (p *progressTracker) wrapReader(r io.Reader) io.Reader {
	return &countingReader{r: r, p: p}
}

// writer returns an io.Writer that counts bytes written (data is discarded).
func (p *progressTracker) writer() io.Writer {
	return &countingWriter{p: p}
}

type countingReader struct {
	r io.Reader
	p *progressTracker
}

func (c *countingReader) Read(buf []byte) (int, error) {
	n, err := c.r.Read(buf)
	c.p.bytes.Add(int64(n))
	return n, err
}

type countingWriter struct {
	p *progressTracker
}

func (c *countingWriter) Write(buf []byte) (int, error) {
	c.p.bytes.Add(int64(len(buf)))
	return len(buf), nil
}
