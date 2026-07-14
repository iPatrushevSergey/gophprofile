package compression

import (
	"compress/gzip"
	"fmt"
	"io"
	"sync"
)

// GzipCompressor implements Compressor using sync.Pool for gzip readers and writers.
type GzipCompressor struct {
	writerPool sync.Pool
	readerPool sync.Pool
}

// NewGzipCompressor returns a GzipCompressor.
func NewGzipCompressor(level int) (*GzipCompressor, error) {
	if _, err := gzip.NewWriterLevel(io.Discard, level); err != nil {
		return nil, fmt.Errorf("gzip compressor level %d: %w", level, err)
	}
	g := &GzipCompressor{}

	g.writerPool = sync.Pool{
		New: func() any {
			w, _ := gzip.NewWriterLevel(io.Discard, level)
			return w
		},
	}
	g.readerPool = sync.Pool{
		New: func() any {
			return new(gzip.Reader)
		},
	}

	return g, nil
}

func (g *GzipCompressor) ContentEncoding() string {
	return "gzip"
}

func (g *GzipCompressor) NewReader(r io.Reader) (io.ReadCloser, error) {
	gzipReader := g.readerPool.Get().(*gzip.Reader)
	if err := gzipReader.Reset(r); err != nil {
		g.readerPool.Put(gzipReader)
		return nil, err
	}
	return &pooledGzipReader{
		Reader: gzipReader,
		pool:   &g.readerPool,
	}, nil
}

func (g *GzipCompressor) NewWriter(w io.Writer) io.WriteCloser {
	gzipWriter := g.writerPool.Get().(*gzip.Writer)
	gzipWriter.Reset(w)
	return &pooledGzipWriter{
		Writer: gzipWriter,
		pool:   &g.writerPool,
	}
}

type pooledGzipReader struct {
	*gzip.Reader
	pool *sync.Pool
}

func (r *pooledGzipReader) Close() error {
	err := r.Reader.Close()
	r.pool.Put(r.Reader)
	return err
}

type pooledGzipWriter struct {
	*gzip.Writer
	pool *sync.Pool
}

func (w *pooledGzipWriter) Close() error {
	err := w.Writer.Close()
	w.pool.Put(w.Writer)
	return err
}
