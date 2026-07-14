// Package compression provides HTTP compression middleware.
package compression

import (
	"io"
	"net/http"
	"strings"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/labstack/echo/v4"
)

// Compressor defines a strategy for compression algorithms (gzip, brotli, etc.).
type Compressor interface {
	ContentEncoding() string
	NewReader(r io.Reader) (io.ReadCloser, error)
	NewWriter(w io.Writer) io.WriteCloser
}

// CompressMiddleware supporting multiple compression strategies.
func CompressMiddleware(log pkgport.Logger, compressors ...Compressor) echo.MiddlewareFunc {
	encodingToCompressor := make(map[string]Compressor)
	for _, compressor := range compressors {
		encodingToCompressor[compressor.ContentEncoding()] = compressor
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Vary", "Accept-Encoding")

			// Decompress Request
			reqEncoding := c.Request().Header.Get("Content-Encoding")
			if reqEncoding != "" {
				if reqCompressor, ok := encodingToCompressor[reqEncoding]; ok {
					compressorReader, err := reqCompressor.NewReader(c.Request().Body)
					if err != nil {
						log.Error("failed to create decompress reader", "error", err, "encoding", reqEncoding)
						return c.NoContent(http.StatusBadRequest)
					}
					c.Request().Body = compressorReader
					defer func() {
						if err := compressorReader.Close(); err != nil {
							log.Warn("failed to close decompress reader", "error", err, "encoding", reqEncoding)
						}
					}()
				}
			}

			// Compress Response
			resEncoding := c.Request().Header.Get("Accept-Encoding")
			var resCompressor Compressor
			for _, compressor := range compressors {
				if strings.Contains(resEncoding, compressor.ContentEncoding()) {
					resCompressor = compressor
					break
				}
			}

			if resCompressor == nil {
				return next(c)
			}

			res := c.Response()
			originalWriter := res.Writer
			writerWithCompressor := &WriterWithCompressor{
				ResponseWriter: res.Writer,
				compressor:     resCompressor,
			}
			res.Writer = writerWithCompressor

			defer func() {
				res.Writer = originalWriter
				if writerWithCompressor.compressorWriter != nil {
					if err := writerWithCompressor.compressorWriter.Close(); err != nil {
						log.Warn(
							"failed to close response compressor writer",
							"error", err,
							"encoding", writerWithCompressor.compressor.ContentEncoding(),
						)
					}
				}
			}()

			return next(c)
		}
	}
}

// WriterWithCompressor creates compressorWriter in WriteHeader only when Content-Type allows compression.
type WriterWithCompressor struct {
	http.ResponseWriter

	compressor       Compressor
	compressorWriter io.WriteCloser

	wroteHeader    bool
	shouldCompress bool
}

func (cw *WriterWithCompressor) Write(data []byte) (int, error) {
	if !cw.wroteHeader {
		cw.WriteHeader(http.StatusOK)
	}
	if cw.shouldCompress && cw.compressorWriter != nil {
		return cw.compressorWriter.Write(data)
	}
	return cw.ResponseWriter.Write(data)
}

func (cw *WriterWithCompressor) WriteHeader(code int) {
	if cw.wroteHeader {
		return
	}
	cw.wroteHeader = true

	contentType := cw.Header().Get("Content-Type")
	if shouldCompress(contentType) {
		cw.shouldCompress = true
		cw.Header().Set("Content-Encoding", cw.compressor.ContentEncoding())
		cw.Header().Del("Content-Length")
		cw.compressorWriter = cw.compressor.NewWriter(cw.ResponseWriter)
	}

	cw.ResponseWriter.WriteHeader(code)
}

func (cw *WriterWithCompressor) WriteString(s string) (int, error) {
	return cw.Write([]byte(s))
}

func shouldCompress(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return false
	}
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "application/xml")
}
