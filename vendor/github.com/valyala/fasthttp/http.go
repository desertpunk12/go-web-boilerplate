package fasthttp

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"os"
	"sync"
	"time"

	"github.com/valyala/bytebufferpool"
)

var (
	requestBodyPoolSizeLimit  = -1
	responseBodyPoolSizeLimit = -1
)

// SetBodySizePoolLimit set the max body size for bodies to be returned to the pool.
// If the body size is larger it will be released instead of put back into the pool for reuse.
func SetBodySizePoolLimit(reqBodyLimit, respBodyLimit int) {
	requestBodyPoolSizeLimit = reqBodyLimit
	responseBodyPoolSizeLimit = respBodyLimit
}

// Request represents HTTP request.
//
// It is forbidden copying Request instances. Create new instances
// and use CopyTo instead.
//
// Request instance MUST NOT be used from concurrently running goroutines.
type Request struct {
	noCopy noCopy

	bodyStream io.Reader
	w          requestBodyWriter
	body       *bytebufferpool.ByteBuffer

	multipartForm         *multipart.Form
	multipartFormBoundary string

	postArgs   Args
	userValues userData

	bodyRaw []byte

	uri URI

	// Request header.
	//
	// Copying Header by value is forbidden. Use pointer to Header instead.
	Header RequestHeader

	// Request timeout. Usually set by DoDeadline or DoTimeout
	// if <= 0, means not set
	timeout time.Duration

	secureErrorLogMessage bool

	// Group bool members in order to reduce Request object size.
	parsedURI      bool
	parsedPostArgs bool

	keepBodyBuffer bool

	// Used by Server to indicate the request was received on a HTTPS endpoint.
	// Client/HostClient shouldn't use this field but should depend on the uri.scheme instead.
	isTLS bool

	// Use Host header (request.Header.SetHost) instead of the host from SetRequestURI, SetHost, or URI().SetHost
	UseHostHeader bool

	// DisableRedirectPathNormalizing disables redirect path normalization when used with DoRedirects.
	//
	// By default redirect path values are normalized, i.e.
	// extra slashes are removed, special characters are encoded.
	DisableRedirectPathNormalizing bool
}

// Response represents HTTP response.
//
// It is forbidden copying Response instances. Create new instances
// and use CopyTo instead.
//
// Response instance MUST NOT be used from concurrently running goroutines.
type Response struct {
	noCopy noCopy

	bodyStream io.Reader

	// Remote TCPAddr from concurrently net.Conn.
	raddr net.Addr
	// Local TCPAddr from concurrently net.Conn.
	laddr net.Addr
	w     responseBodyWriter
	body  *bytebufferpool.ByteBuffer

	bodyRaw []byte

	// Response header.
	//
	// Copying Header by value is forbidden. Use pointer to Header instead.
	Header ResponseHeader

	// Flush headers as soon as possible without waiting for first body bytes.
	// Relevant for bodyStream only.
	ImmediateHeaderFlush bool

	// StreamBody enables response body streaming.
	// Use SetBodyStream to set the body stream.
	StreamBody bool

	// Response.Read() skips reading body if set to true.
	// Use it for reading HEAD responses.
	//
	// Response.Write() skips writing body if set to true.
	// Use it for writing HEAD responses.
	SkipBody bool

	keepBodyBuffer        bool
	secureErrorLogMessage bool
}

// SetHost sets host for the request.
func (req *Request) SetHost(host string) {
	req.URI().SetHost(host)
}

// SetHostBytes sets host for the request.
func (req *Request) SetHostBytes(host []byte) {
	req.URI().SetHostBytes(host)
}

// Host returns the host for the given request.
func (req *Request) Host() []byte {
	return req.URI().Host()
}

// SetRequestURI sets RequestURI.
func (req *Request) SetRequestURI(requestURI string) {
	req.Header.SetRequestURI(requestURI)
	req.parsedURI = false
}

// SetRequestURIBytes sets RequestURI.
func (req *Request) SetRequestURIBytes(requestURI []byte) {
	req.Header.SetRequestURIBytes(requestURI)
	req.parsedURI = false
}

// RequestURI returns request's URI.
func (req *Request) RequestURI() []byte {
	if req.parsedURI {
		requestURI := req.uri.RequestURI()
		req.SetRequestURIBytes(requestURI)
	}
	return req.Header.RequestURI()
}

// StatusCode returns response status code.
func (resp *Response) StatusCode() int {
	return resp.Header.StatusCode()
}

// SetStatusCode sets response status code.
func (resp *Response) SetStatusCode(statusCode int) {
	resp.Header.SetStatusCode(statusCode)
}

// ConnectionClose returns true if 'Connection: close' header is set.
func (resp *Response) ConnectionClose() bool {
	return resp.Header.ConnectionClose()
}

// SetConnectionClose sets 'Connection: close' header.
func (resp *Response) SetConnectionClose() {
	resp.Header.SetConnectionClose()
}

// ConnectionClose returns true if 'Connection: close' header is set.
func (req *Request) ConnectionClose() bool {
	return req.Header.ConnectionClose()
}

// SetConnectionClose sets 'Connection: close' header.
func (req *Request) SetConnectionClose() {
	req.Header.SetConnectionClose()
}

// GetTimeOut retrieves the timeout duration set for the Request.
//
// This method returns a time.Duration that determines how long the request
// can wait before it times out. In the default use case, the timeout applies
// to the entire request lifecycle, including both receiving the response
// headers and the response body.
func (req *Request) GetTimeOut() time.Duration {
	return req.timeout
}

// SendFile registers file on the given path to be used as response body
// when Write is called.
//
// Note that SendFile doesn't set Content-Type, so set it yourself
// with Header.SetContentType.
func (resp *Response) SendFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	fileInfo, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	size64 := fileInfo.Size()
	size := int(size64)
	if int64(size) != size64 {
		size = -1
	}

	resp.Header.SetLastModified(fileInfo.ModTime())
	resp.SetBodyStream(f, size)
	return nil
}

// SetBodyStream sets request body stream and, optionally body size.
//
// If bodySize is >= 0, then the bodyStream must provide exactly bodySize bytes
// before returning io.EOF.
//
// If bodySize < 0, then bodyStream is read until io.EOF.
//
// bodyStream.Close() is called after finishing reading all body data
// if it implements io.Closer.
//
// Note that GET and HEAD requests cannot have body.
//
// See also SetBodyStreamWriter.
func (req *Request) SetBodyStream(bodyStream io.Reader, bodySize int) {
	req.ResetBody()
	req.bodyStream = bodyStream
	req.Header.SetContentLength(bodySize)
}

// SetBodyStream sets response body stream and, optionally body size.
//
// If bodySize is >= 0, then the bodyStream must provide exactly bodySize bytes
// before returning io.EOF.
//
// If bodySize < 0, then bodyStream is read until io.EOF.
//
// bodyStream.Close() is called after finishing reading all body data
// if it implements io.Closer.
//
// See also SetBodyStreamWriter.
func (resp *Response) SetBodyStream(bodyStream io.Reader, bodySize int) {
	resp.ResetBody()
	resp.bodyStream = bodyStream
	resp.Header.SetContentLength(bodySize)
}

// IsBodyStream returns true if body is set via SetBodyStream*.
func (req *Request) IsBodyStream() bool {
	return req.bodyStream != nil
}

// IsBodyStream returns true if body is set via SetBodyStream*.
func (resp *Response) IsBodyStream() bool {
	return resp.bodyStream != nil
}

// SetBodyStreamWriter registers the given sw for populating request body.
//
// This function may be used in the following cases:
//
//   - if request body is too big (more than 10MB).
//   - if request body is streamed from slow external sources.
//   - if request body must be streamed to the server in chunks
//     (aka `http client push` or `chunked transfer-encoding`).
//
// Note that GET and HEAD requests cannot have body.
//
// See also SetBodyStream.
func (req *Request) SetBodyStreamWriter(sw StreamWriter) {
	sr := NewStreamReader(sw)
	req.SetBodyStream(sr, -1)
}

// SetBodyStreamWriter registers the given sw for populating response body.
//
// This function may be used in the following cases:
//
//   - if response body is too big (more than 10MB).
//   - if response body is streamed from slow external sources.
//   - if response body must be streamed to the client in chunks
//     (aka `http server push` or `chunked transfer-encoding`).
//
// See also SetBodyStream.
func (resp *Response) SetBodyStreamWriter(sw StreamWriter) {
	sr := NewStreamReader(sw)
	resp.SetBodyStream(sr, -1)
}

// BodyWriter returns writer for populating response body.
//
// If used inside RequestHandler, the returned writer must not be used
// after returning from RequestHandler. Use RequestCtx.Write
// or SetBodyStreamWriter in this case.
func (resp *Response) BodyWriter() io.Writer {
	resp.w.r = resp
	return &resp.w
}

// BodyStream returns io.Reader.
//
// You must CloseBodyStream or ReleaseRequest after you use it.
func (req *Request) BodyStream() io.Reader {
	return req.bodyStream
}

func (req *Request) CloseBodyStream() error {
	return req.closeBodyStream()
}

// BodyStream returns io.Reader.
//
// You must CloseBodyStream or ReleaseResponse after you use it.
func (resp *Response) BodyStream() io.Reader {
	return resp.bodyStream
}

func (resp *Response) CloseBodyStream() error {
	return resp.closeBodyStream(nil)
}

type ReadCloserWithError interface {
	io.Reader
	CloseWithError(err error) error
}

type closeReader struct {
	io.Reader
	closeFunc func(err error) error
}

func newCloseReaderWithError(r io.Reader, closeFunc func(err error) error) ReadCloserWithError {
	if r == nil {
		panic(`BUG: reader is nil`)
	}
	return &closeReader{Reader: r, closeFunc: closeFunc}
}

func (c *closeReader) CloseWithError(err error) error {
	if c.closeFunc == nil {
		return nil
	}
	return c.closeFunc(err)
}

// BodyWriter returns writer for populating request body.
func (req *Request) BodyWriter() io.Writer {
	req.w.r = req
	return &req.w
}

type responseBodyWriter struct {
	r *Response
}

func (w *responseBodyWriter) Write(p []byte) (int, error) {
	w.r.AppendBody(p)
	return len(p), nil
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.r.AppendBodyString(s)
	return len(s), nil
}

type requestBodyWriter struct {
	r *Request
}

func (w *requestBodyWriter) Write(p []byte) (int, error) {
	w.r.AppendBody(p)
	return len(p), nil
}

func (w *requestBodyWriter) WriteString(s string) (int, error) {
	w.r.AppendBodyString(s)
	return len(s), nil
}

func (resp *Response) ParseNetConn(conn net.Conn) {
	resp.raddr = conn.RemoteAddr()
	resp.laddr = conn.LocalAddr()
}

// RemoteAddr returns the remote network address. The Addr returned is shared
// by all invocations of RemoteAddr, so do not modify it.
func (resp *Response) RemoteAddr() net.Addr {
	return resp.raddr
}

// LocalAddr returns the local network address. The Addr returned is shared
// by all invocations of LocalAddr, so do not modify it.
func (resp *Response) LocalAddr() net.Addr {
	return resp.laddr
}

// Body returns response body.
//
// The returned value is valid until the response is released,
// either though ReleaseResponse or your request handler returning.
// Do not store references to returned value. Make copies instead.
func (resp *Response) Body() []byte {
	if resp.bodyStream != nil {
		bodyBuf := resp.bodyBuffer()
		bodyBuf.Reset()
		_, err := copyZeroAlloc(bodyBuf, resp.bodyStream)
		resp.closeBodyStream(err) //nolint:errcheck
		if err != nil {
			bodyBuf.SetString(err.Error())
		}
	}
	return resp.bodyBytes()
}

func (resp *Response) bodyBytes() []byte {
	if resp.bodyRaw != nil {
		return resp.bodyRaw
	}
	if resp.body == nil {
		return nil
	}
	return resp.body.B
}

func (req *Request) bodyBytes() []byte {
	if req.bodyRaw != nil {
		return req.bodyRaw
	}
	if req.bodyStream != nil {
		bodyBuf := req.bodyBuffer()
		bodyBuf.Reset()
		_, err := copyZeroAlloc(bodyBuf, req.bodyStream)
		req.closeBodyStream() //nolint:errcheck
		if err != nil {
			bodyBuf.SetString(err.Error())
		}
	}
	if req.body == nil {
		return nil
	}
	return req.body.B
}

func (resp *Response) bodyBuffer() *bytebufferpool.ByteBuffer {
	if resp.body == nil {
		resp.body = responseBodyPool.Get()
	}
	resp.bodyRaw = nil
	return resp.body
}

func (req *Request) bodyBuffer() *bytebufferpool.ByteBuffer {
	if req.body == nil {
		req.body = requestBodyPool.Get()
	}
	req.bodyRaw = nil
	return req.body
}

var (
	responseBodyPool bytebufferpool.Pool
	requestBodyPool  bytebufferpool.Pool
)

// BodyGunzip returns un-gzipped body data.
//
// This method may be used if the request header contains
// 'Content-Encoding: gzip' for reading un-gzipped body.
// Use Body for reading gzipped request body.
func (req *Request) BodyGunzip() ([]byte, error) {
	return gunzipData(req.Body())
}

// BodyGunzip returns un-gzipped body data.
//
// This method may be used if the response header contains
// 'Content-Encoding: gzip' for reading un-gzipped body.
// Use Body for reading gzipped response body.
func (resp *Response) BodyGunzip() ([]byte, error) {
	return gunzipData(resp.Body())
}

func gunzipData(p []byte) ([]byte, error) {
	var bb bytebufferpool.ByteBuffer
	_, err := WriteGunzip(&bb, p)
	if err != nil {
		return nil, err
	}
	return bb.B, nil
}

// BodyUnbrotli returns un-brotlied body data.
//
// This method may be used if the request header contains
// 'Content-Encoding: br' for reading un-brotlied body.
// Use Body for reading brotlied request body.
func (req *Request) BodyUnbrotli() ([]byte, error) {
	return unBrotliData(req.Body())
}

// BodyUnbrotli returns un-brotlied body data.
//
// This method may be used if the response header contains
// 'Content-Encoding: br' for reading un-brotlied body.
// Use Body for reading brotlied response body.
func (resp *Response) BodyUnbrotli() ([]byte, error) {
	return unBrotliData(resp.Body())
}

func unBrotliData(p []byte) ([]byte, error) {
	var bb bytebufferpool.ByteBuffer
	_, err := WriteUnbrotli(&bb, p)
	if err != nil {
		return nil, err
	}
	return bb.B, nil
}

// BodyInflate returns inflated body data.
//
// This method may be used if the response header contains
// 'Content-Encoding: deflate' for reading inflated request body.
// Use Body for reading deflated request body.
func (req *Request) BodyInflate() ([]byte, error) {
	return inflateData(req.Body())
}

// BodyInflate returns inflated body data.
//
// This method may be used if the response header contains
// 'Content-Encoding: deflate' for reading inflated response body.
// Use Body for reading deflated response body.
func (resp *Response) BodyInflate() ([]byte, error) {
	return inflateData(resp.Body())
}

func (ctx *RequestCtx) RequestBodyStream() io.Reader {
	return ctx.Request.bodyStream
}

func (req *Request) BodyUnzstd() ([]byte, error) {
	return unzstdData(req.Body())
}

func (resp *Response) BodyUnzstd() ([]byte, error) {
	return unzstdData(resp.Body())
}

func unzstdData(p []byte) ([]byte, error) {
	var bb bytebufferpool.ByteBuffer
	_, err := WriteUnzstd(&bb, p)
	if err != nil {
		return nil, err
	}
	return bb.B, nil
}

func inflateData(p []byte) ([]byte, error) {
	var bb bytebufferpool.ByteBuffer
	_, err := WriteInflate(&bb, p)
	if err != nil {
		return nil, err
	}
	return bb.B, nil
}

var ErrContentEncodingUnsupported = errors.New("unsupported Content-Encoding")

// BodyUncompressed returns body data and if needed decompress it from gzip, deflate or Brotli.
//
// This method may be used if the response header contains
// 'Content-Encoding' for reading uncompressed request body.
// Use Body for reading the raw request body.
func (req *Request) BodyUncompressed() ([]byte, error) {
	switch string(req.Header.ContentEncoding()) {
	case "":
		return req.Body(), nil
	case "deflate":
		return req.BodyInflate()
	case "gzip":
		return req.BodyGunzip()
	case "br":
		return req.BodyUnbrotli()
	case "zstd":
		return req.BodyUnzstd()
	default:
		return nil, ErrContentEncodingUnsupported
	}
}

// BodyUncompressed returns body data and if needed decompress it from gzip, deflate or Brotli.
//
// This method may be used if the response header contains
// 'Content-Encoding' for reading uncompressed response body.
// Use Body for reading the raw response body.
func (resp *Response) BodyUncompressed() ([]byte, error) {
	switch string(resp.Header.ContentEncoding()) {
	case "":
		return resp.Body(), nil
	case "deflate":
		return resp.BodyInflate()
	case "gzip":
		return resp.BodyGunzip()
	case "br":
		return resp.BodyUnbrotli()
	case "zstd":
		return resp.BodyUnzstd()
	default:
		return nil, ErrContentEncodingUnsupported
	}
}

// BodyWriteTo writes request body to w.
func (req *Request) BodyWriteTo(w io.Writer) error {
	if req.bodyStream != nil {
		_, err := copyZeroAlloc(w, req.bodyStream)
		req.closeBodyStream() //nolint:errcheck
		return err
	}
	if req.onlyMultipartForm() {
		return WriteMultipartForm(w, req.multipartForm, req.multipartFormBoundary)
	}
	_, err := w.Write(req.bodyBytes())
	return err
}

// BodyWriteTo writes response body to w.
func (resp *Response) BodyWriteTo(w io.Writer) error {
	if resp.bodyStream != nil {
		_, err := copyZeroAlloc(w, resp.bodyStream)
		resp.closeBodyStream(err) //nolint:errcheck
		return err
	}
	_, err := w.Write(resp.bodyBytes())
	return err
}

// AppendBody appends p to response body.
//
// It is safe re-using p after the function returns.
func (resp *Response) AppendBody(p []byte) {
	resp.closeBodyStream(nil)  //nolint:errcheck
	resp.bodyBuffer().Write(p) //nolint:errcheck
}

// AppendBodyString appends s to response body.
func (resp *Response) AppendBodyString(s string) {
	resp.closeBodyStream(nil)        //nolint:errcheck
	resp.bodyBuffer().WriteString(s) //nolint:errcheck
}

// SetBody sets response body.
//
// It is safe re-using body argument after the function returns.
func (resp *Response) SetBody(body []byte) {
	resp.closeBodyStream(nil) //nolint:errcheck
	bodyBuf := resp.bodyBuffer()
	bodyBuf.Reset()
	bodyBuf.Write(body) //nolint:errcheck
}

// SetBodyString sets response body.
func (resp *Response) SetBodyString(body string) {
	resp.closeBodyStream(nil) //nolint:errcheck
	bodyBuf := resp.bodyBuffer()
	bodyBuf.Reset()
	bodyBuf.WriteString(body) //nolint:errcheck
}

// ResetBody resets response body.
func (resp *Response) ResetBody() {
	resp.bodyRaw = nil
	resp.closeBodyStream(nil) //nolint:errcheck
	if resp.body != nil {
		if resp.keepBodyBuffer {
			resp.body.Reset()
		} else {
			responseBodyPool.Put(resp.body)
			resp.body = nil
		}
	}
}

// SetBodyRaw sets response body, but without copying it.
//
// From this point onward the body argument must not be changed.
func (resp *Response) SetBodyRaw(body []byte) {
	resp.ResetBody()
	resp.bodyRaw = body
}

// SetBodyRaw sets response body, but without copying it.
//
// From this point onward the body argument must not be changed.
func (req *Request) SetBodyRaw(body []byte) {
	req.ResetBody()
	req.bodyRaw = body
}

// ReleaseBody retires the response body if it is greater than "size" bytes.
//
// This permits GC to reclaim the large buffer.  If used, must be before
// ReleaseResponse.
//
// Use this method only if you really understand how it works.
// The majority of workloads don't need this method.
func (resp *Response) ReleaseBody(size int) {
	resp.bodyRaw = nil
	if resp.body == nil {
		return
	}
	if cap(resp.body.B) > size {
		resp.closeBodyStream(nil) //nolint:errcheck
		resp.body = nil
	}
}

// ReleaseBody retires the request body if it is greater than "size" bytes.
//
// This permits GC to reclaim the large buffer.  If used, must be before
// ReleaseRequest.
//
// Use this method only if you really understand how it works.
// The majority of workloads don't need this method.
func (req *Request) ReleaseBody(size int) {
	req.bodyRaw = nil
	if req.body == nil {
		return
	}
	if cap(req.body.B) > size {
		req.closeBodyStream() //nolint:errcheck
		req.body = nil
	}
}

// SwapBody swaps response body with the given body and returns
// the previous response body.
//
// It is forbidden to use the body passed to SwapBody after
// the function returns.
func (resp *Response) SwapBody(body []byte) []byte {
	bb := resp.bodyBuffer()

	if resp.bodyStream != nil {
		bb.Reset()
		_, err := copyZeroAlloc(bb, resp.bodyStream)
		resp.closeBodyStream(err) //nolint:errcheck
		if err != nil {
			bb.Reset()
			bb.SetString(err.Error())
		}
	}

	resp.bodyRaw = nil

	oldBody := bb.B
	bb.B = body
	return oldBody
}

// SwapBody swaps request body with the given body and returns
// the previous request body.
//
// It is forbidden to use the body passed to SwapBody after
// the function returns.
func (req *Request) SwapBody(body []byte) []byte {
	bb := req.bodyBuffer()

	if req.bodyStream != nil {
		bb.Reset()
		_, err := copyZeroAlloc(bb, req.bodyStream)
		req.closeBodyStream() //nolint:errcheck
		if err != nil {
			bb.Reset()
			bb.SetString(err.Error())
		}
	}

	req.bodyRaw = nil

	oldBody := bb.B
	bb.B = body
	return oldBody
}

// Body returns request body.
//
// The returned value is valid until the request is released,
// either though ReleaseRequest or your request handler returning.
// Do not store references to returned value. Make copies instead.
func (req *Request) Body() []byte {
	if req.bodyRaw != nil {
		return req.bodyRaw
	} else if req.onlyMultipartForm() {
		body, err := marshalMultipartForm(req.multipartForm, req.multipartFormBoundary)
		if err != nil {
			return []byte(err.Error())
		}
		return body
	}
	return req.bodyBytes()
}

// AppendBody appends p to request body.
//
// It is safe re-using p after the function returns.
func (req *Request) AppendBody(p []byte) {
	req.RemoveMultipartFormFiles()
	req.closeBodyStream()     //nolint:errcheck
	req.bodyBuffer().Write(p) //nolint:errcheck
}

// AppendBodyString appends s to request body.
func (req *Request) AppendBodyString(s string) {
	req.RemoveMultipartFormFiles()
	req.closeBodyStream()           //nolint:errcheck
	req.bodyBuffer().WriteString(s) //nolint:errcheck
}

// SetBody sets request body.
//
// It is safe re-using body argument after the function returns.
func (req *Request) SetBody(body []byte) {
	req.RemoveMultipartFormFiles()
	req.closeBodyStream() //nolint:errcheck
	req.bodyBuffer().Set(body)
}

// SetBodyString sets request body.
func (req *Request) SetBodyString(body string) {
	req.RemoveMultipartFormFiles()
	req.closeBodyStream() //nolint:errcheck
	req.bodyBuffer().SetString(body)
}

// ResetBody resets request body.
func (req *Request) ResetBody() {
	req.bodyRaw = nil
	req.RemoveMultipartFormFiles()
	req.closeBodyStream() //nolint:errcheck
	if req.body != nil {
		if req.keepBodyBuffer {
			req.body.Reset()
		} else {
			requestBodyPool.Put(req.body)
			req.body = nil
		}
	}
}

// CopyTo copies req contents to dst except of body stream.
func (req *Request) CopyTo(dst *Request) {
	req.copyToSkipBody(dst)
	switch {
	case req.bodyRaw != nil:
		dst.bodyRaw = append(dst.bodyRaw[:0], req.bodyRaw...)
		if dst.body != nil {
			dst.body.Reset()
		}
	case req.body != nil:
		dst.bodyBuffer().Set(req.body.B)
	case dst.body != nil:
		dst.body.Reset()
	}
}

func (req *Request) copyToSkipBody(dst *Request) {
	dst.Reset()
	req.Header.CopyTo(&dst.Header)

	req.uri.CopyTo(&dst.uri)
	dst.parsedURI = req.parsedURI

	req.postArgs.CopyTo(&dst.postArgs)
	dst.parsedPostArgs = req.parsedPostArgs
	dst.isTLS = req.isTLS

	dst.UseHostHeader = req.UseHostHeader

	// do not copy multipartForm - it will be automatically
	// re-created on the first call to MultipartForm.
}

// CopyTo copies resp contents to dst except of body stream.
func (resp *Response) CopyTo(dst *Response) {
	resp.copyToSkipBody(dst)
	switch {
	case resp.bodyRaw != nil:
		dst.bodyRaw = append(dst.bodyRaw, resp.bodyRaw...)
		if dst.body != nil {
			dst.body.Reset()
		}
	case resp.body != nil:
		dst.bodyBuffer().Set(resp.body.B)
	case dst.body != nil:
		dst.body.Reset()
	}
}

func (resp *Response) copyToSkipBody(dst *Response) {
	dst.Reset()
	resp.Header.CopyTo(&dst.Header)
	dst.SkipBody = resp.SkipBody
	dst.raddr = resp.raddr
	dst.laddr = resp.laddr
}

func swapRequestBody(a, b *Request) {
	a.body, b.body = b.body, a.body
	a.bodyRaw, b.bodyRaw = b.bodyRaw, a.bodyRaw
	a.bodyStream, b.bodyStream = b.bodyStream, a.bodyStream

	// This code assumes that if a requestStream was swapped the headers are also swapped or copied.
	if rs, ok := a.bodyStream.(*requestStream); ok {
		rs.header = &a.Header
	}
	if rs, ok := b.bodyStream.(*requestStream); ok {
		rs.header = &b.Header
	}
}

func swapResponseBody(a, b *Response) {
	a.body, b.body = b.body, a.body
	a.bodyRaw, b.bodyRaw = b.bodyRaw, a.bodyRaw
	a.bodyStream, b.bodyStream = b.bodyStream, a.bodyStream
}

// URI returns request URI.
func (req *Request) URI() *URI {
	req.parseURI() //nolint:errcheck
	return &req.uri
}

// SetURI initializes request URI.
// Use this method if a single URI may be reused across multiple requests.
// Otherwise, you can just use SetRequestURI() and it will be parsed as new URI.
// The URI is copied and can be safely modified later.
func (req *Request) SetURI(newURI *URI) {
	if newURI != nil {
		newURI.CopyTo(&req.uri)
		req.parsedURI = true
		return
	}
	req.uri.Reset()
	req.parsedURI = false
}

func (req *Request) parseURI() error {
	if req.parsedURI {
		return nil
	}
	req.parsedURI = true

	return req.uri.parse(req.Header.Host(), req.Header.RequestURI(), req.isTLS)
}

// PostArgs returns POST arguments.
func (req *Request) PostArgs() *Args {
	req.parsePostArgs()
	return &req.postArgs
}

func (req *Request) parsePostArgs() {
	if req.parsedPostArgs {
		return
	}
	req.parsedPostArgs = true

	if !bytes.HasPrefix(req.Header.ContentType(), strPostArgsContentType) {
		return
	}
	req.postArgs.ParseBytes(req.bodyBytes())
}

// ErrNoMultipartForm means that the request's Content-Type
// isn't 'multipart/form-data'.
var ErrNoMultipartForm = errors.New("request Content-Type has bad boundary or is not multipart/form-data")

// MultipartForm returns request's multipart form.
//
// Returns ErrNoMultipartForm if request's Content-Type
// isn't 'multipart/form-data'.
//
// RemoveMultipartFormFiles must be called after returned multipart form
// is processed.
func (req *Request) MultipartForm() (*multipart.Form, error) {
	if req.multipartForm != nil {
		return req.multipartForm, nil
	}

	req.multipartFormBoundary = string(req.Header.MultipartFormBoundary())
	if req.multipartFormBoundary == "" {
		return nil, ErrNoMultipartForm
	}

	var err error
	ce := req.Header.peek(strContentEncoding)

	if req.bodyStream != nil {
		bodyStream := req.bodyStream
		if bytes.Equal(ce, strGzip) {
			// Do not care about memory usage here.
			if bodyStream, err = gzip.NewReader(bodyStream); err != nil {
				return nil, fmt.Errorf("cannot gunzip request body: %w", err)
			}
		} else if len(ce) > 0 {
			return nil, fmt.Errorf("unsupported Content-Encoding: %q", ce)
		}

		mr := multipart.NewReader(bodyStream, req.multipartFormBoundary)
		req.multipartForm, err = mr.ReadForm(8 * 1024)
		if err != nil {
			return nil, fmt.Errorf("cannot read multipart/form-data body: %w", err)
		}
	} else {
		body := req.bodyBytes()
		if bytes.Equal(ce, strGzip) {
			// Do not care about memory usage here.
			if body, err = AppendGunzipBytes(nil, body); err != nil {
				return nil, fmt.Errorf("cannot gunzip request body: %w", err)
			}
		} else if len(ce) > 0 {
			return nil, fmt.Errorf("unsupported Content-Encoding: %q", ce)
		}

		req.multipartForm, err = readMultipartForm(bytes.NewReader(body), req.multipartFormBoundary, len(body), len(body))
		if err != nil {
			return nil, err
		}
	}

	return req.multipartForm, nil
}

func marshalMultipartForm(f *multipart.Form, boundary string) ([]byte, error) {
	var buf bytebufferpool.ByteBuffer
	if err := WriteMultipartForm(&buf, f, boundary); err != nil {
		return nil, err
	}
	return buf.B, nil
}

// WriteMultipartForm writes the given multipart form f with the given
// boundary to w.
func WriteMultipartForm(w io.Writer, f *multipart.Form, boundary string) error {
	// Do not care about memory allocations here, since multipart
	// form processing is slow.
	if boundary == "" {
		return errors.New("form boundary cannot be empty")
	}

	mw := multipart.NewWriter(w)
	if err := mw.SetBoundary(boundary); err != nil {
		return fmt.Errorf("cannot use form boundary %q: %w", boundary, err)
	}

	// marshal values
	for k, vv := range f.Value {
		for _, v := range vv {
			if err := mw.WriteField(k, v); err != nil {
				return fmt.Errorf("cannot write form field %q value %q: %w", k, v, err)
			}
		}
	}

	// marshal files
	for k, fvv := range f.File {
		for _, fv := range fvv {
			vw, err := mw.CreatePart(fv.Header)
			if err != nil {
				return fmt.Errorf("cannot create form file %q (%q): %w", k, fv.Filename, err)
			}
			fh, err := fv.Open()
			if err != nil {
				return fmt.Errorf("cannot open form file %q (%q): %w", k, fv.Filename, err)
			}
			if _, err = copyZeroAlloc(vw, fh); err != nil {
				_ = fh.Close()
				return fmt.Errorf("error when copying form file %q (%q): %w", k, fv.Filename, err)
			}
			if err = fh.Close(); err != nil {
				return fmt.Errorf("cannot close form file %q (%q): %w", k, fv.Filename, err)
			}
		}
	}

	if err := mw.Close(); err != nil {
		return fmt.Errorf("error when closing multipart form writer: %w", err)
	}

	return nil
}

func readMultipartForm(r io.Reader, boundary string, size, maxInMemoryFileSize int) (*multipart.Form, error) {
	// Do not care about memory allocations here, since they are tiny
	// compared to multipart data (aka multi-MB files) usually sent
	// in multipart/form-data requests.

	if size <= 0 {
		return nil, fmt.Errorf("form size must be greater than 0. Given %d", size)
	}
	lr := io.LimitReader(r, int64(size))
	mr := multipart.NewReader(lr, boundary)
	f, err := mr.ReadForm(int64(maxInMemoryFileSize))
	if err != nil {
		return nil, fmt.Errorf("cannot read multipart/form-data body: %w", err)
	}
	return f, nil
}

// Reset clears request contents.
func (req *Request) Reset() {
	req.userValues.Reset() // it should be at the top, since some values might implement io.Closer interface
	if requestBodyPoolSizeLimit >= 0 && req.body != nil {
		req.ReleaseBody(requestBodyPoolSizeLimit)
	}
	req.Header.Reset()
	req.resetSkipHeader()
	req.timeout = 0
	req.UseHostHeader = false
	req.DisableRedirectPathNormalizing = false
}

func (req *Request) resetSkipHeader() {
	req.ResetBody()
	req.uri.Reset()
	req.parsedURI = false
	req.postArgs.Reset()
	req.parsedPostArgs = false
	req.isTLS = false
}

// RemoveMultipartFormFiles removes multipart/form-data temporary files
// associated with the request.
func (req *Request) RemoveMultipartFormFiles() {
	if req.multipartForm != nil {
		// Do not check for error, since these files may be deleted or moved
		// to new places by user code.
		req.multipartForm.RemoveAll() //nolint:errcheck
		req.multipartForm = nil
	}
	req.multipartFormBoundary = ""
}

// Reset clears response contents.
func (resp *Response) Reset() {
	if responseBodyPoolSizeLimit >= 0 && resp.body != nil {
		resp.ReleaseBody(responseBodyPoolSizeLimit)
	}
	resp.resetSkipHeader()
	resp.Header.Reset()
	resp.SkipBody = false
	resp.raddr = nil
	resp.laddr = nil
	resp.ImmediateHeaderFlush = false
	resp.StreamBody = false
}

func (resp *Response) resetSkipHeader() {
	resp.ResetBody()
}

// Read reads request (including body) from the given r.
//
// RemoveMultipartFormFiles or Reset must be called after
// reading multipart/form-data request in order to delete temporarily
// uploaded files.
//
// If MayContinue returns true, the caller must:
//
//   - Either send StatusExpectationFailed response if request headers don't
//     satisfy the caller.
//   - Or send StatusContinue response before reading request body
//     with ContinueReadBody.
//   - Or close the connection.
//
// io.EOF is returned if r is closed before reading the first header byte.
func (req *Request) Read(r *bufio.Reader) error {
	return req.ReadLimitBody(r, 0)
}

const defaultMaxInMemoryFileSize = 16 * 1024 * 1024

// ErrGetOnly is returned when server expects only GET requests,
// but some other type of request came (Server.GetOnly option is true).
var ErrGetOnly = errors.New("non-GET request received")

// ReadLimitBody reads request from the given r, limiting the body size.
//
// If maxBodySize > 0 and the body size exceeds maxBodySize,
// then ErrBodyTooLarge is returned.
//
// RemoveMultipartFormFiles or Reset must be called after
// reading multipart/form-data request in order to delete temporarily
// uploaded files.
//
// If MayContinue returns true, the caller must:
//
//   - Either send StatusExpectationFailed response if request headers don't
//     satisfy the caller.
//   - Or send StatusContinue response before reading request body
//     with ContinueReadBody.
//   - Or close the connection.
//
// io.EOF is returned if r is closed before reading the first header byte.
func (req *Request) ReadLimitBody(r *bufio.Reader, maxBodySize int) error {
	req.resetSkipHeader()
	if err := req.Header.Read(r); err != nil {
		return err
	}

	return req.readLimitBody(r, maxBodySize, false, true)
}

func (req *Request) readLimitBody(r *bufio.Reader, maxBodySize int, getOnly, preParseMultipartForm bool) error {
	// Do not reset the request here - the caller must reset it before
	// calling this method.

	if getOnly && !req.Header.IsGet() && !req.Header.IsHead() {
		return ErrGetOnly
	}

	if req.MayContinue() {
		// 'Expect: 100-continue' header found. Let the caller deciding
		// whether to read request body or
		// to return StatusExpectationFailed.
		return nil
	}

	return req.ContinueReadBody(r, maxBodySize, preParseMultipartForm)
}

func (req *Request) readBodyStream(r *bufio.Reader, maxBodySize int, getOnly, preParseMultipartForm bool) error {
	// Do not reset the request here - the caller must reset it before
	// calling this method.

	if getOnly && !req.Header.IsGet() && !req.Header.IsHead() {
		return ErrGetOnly
	}

	if req.MayContinue() {
		// 'Expect: 100-continue' header found. Let the caller deciding
		// whether to read request body or
		// to return StatusExpectationFailed.
		return nil
	}

	return req.ContinueReadBodyStream(r, maxBodySize, preParseMultipartForm)
}

// MayContinue returns true if the request contains
// 'Expect: 100-continue' header.
//
// The caller must do one of the following actions if MayContinue returns true:
//
//   - Either send StatusExpectationFailed response if request headers don't
//     satisfy the caller.
//   - Or send StatusContinue response before reading request body
//     with ContinueReadBody.
//   - Or close the connection.
func (req *Request) MayContinue() bool {
	return bytes.Equal(req.Header.peek(strExpect), str100Continue)
}

// ContinueReadBody reads request body if request header contains
// 'Expect: 100-continue'.
//
// The caller must send StatusContinue response before calling this method.
//
// If maxBodySize > 0 and the body size exceeds maxBodySize,
// then ErrBodyTooLarge is returned.
func (req *Request) ContinueReadBody(r *bufio.Reader, maxBodySize int, preParseMultipartForm ...bool) error {
	var err error
	contentLength := req.Header.ContentLength()
	if contentLength > 0 {
		if maxBodySize > 0 && contentLength > maxBodySize {
			return ErrBodyTooLarge
		}

		if len(preParseMultipartForm) == 0 || preParseMultipartForm[0] {
			// Pre-read multipart form data of known length.
			// This way we limit memory usage for large file uploads, since their contents
			// is streamed into temporary files if file size exceeds defaultMaxInMemoryFileSize.
			req.multipartFormBoundary = string(req.Header.MultipartFormBoundary())
			if req.multipartFormBoundary != "" && len(req.Header.peek(strContentEncoding)) == 0 {
				req.multipartForm, err = readMultipartForm(r, req.multipartFormBoundary, contentLength, defaultMaxInMemoryFileSize)
				if err != nil {
					req.Reset()
				}
				return err
			}
		}
	}

	if contentLength == -2 {
		// identity body has no sense for http requests, since
		// the end of body is determined by connection close.
		// So just ignore request body for requests without
		// 'Content-Length' and 'Transfer-Encoding' headers.
		// refer to https://tools.ietf.org/html/rfc7230#section-3.3.2
		if !req.Header.ignoreBody() {
			req.Header.SetContentLength(0)
		}
		return nil
	}

	if err = req.ReadBody(r, contentLength, maxBodySize); err != nil {
		return err
	}

	if contentLength == -1 {
		err = req.Header.ReadTrailer(r)
		if err != nil && err != io.EOF {
			return err
		}
	}
	return nil
}

// ReadBody reads request body from the given r, limiting the body size.
//
// If maxBodySize > 0 and the body size exceeds maxBodySize,
// then ErrBodyTooLarge is returned.
func (req *Request) ReadBody(r *bufio.Reader, contentLength, maxBodySize int) (err error) {
	bodyBuf := req.bodyBuffer()
	bodyBuf.Reset()

	switch {
	case contentLength >= 0:
		bodyBuf.B, err = readBody(r, contentLength, maxBodySize, bodyBuf.B)
	case contentLength == -1:
		bodyBuf.B, err = readBodyChunked(r, maxBodySize, bodyBuf.B)
		if err == nil && len(bodyBuf.B) == 0 {
			req.Header.SetContentLength(0)
		}
	default:
		bodyBuf.B, err = readBodyIdentity(r, maxBodySize, bodyBuf.B)
		req.Header.SetContentLength(len(bodyBuf.B))
	}

	if err != nil {
		req.Reset()
		return err
	}
	return nil
}

// ContinueReadBodyStream reads request body if request header contains
// 'Expect: 100-continue'.
//
// The caller must send StatusContinue response before calling this method.
//
// If maxBodySize > 0 and the body size exceeds maxBodySize,
// then ErrBodyTooLarge is returned.
func (req *Request) ContinueReadBodyStream(r *bufio.Reader, maxBodySize int, preParseMultipartForm ...bool) error {
	var err error
	contentLength := req.Header.ContentLength()
	if contentLength > 0 {
		if len(preParseMultipartForm) == 0 || preParseMultipartForm[0] {
			// Pre-read multipart form data of known length.
			// This way we limit memory usage for large file uploads, since their contents
			// is streamed into temporary files if file size exceeds defaultMaxInMemoryFileSize.
			req.multipartFormBoundary = b2s(req.Header.MultipartFormBoundary())
			if req.multipartFormBoundary != "" && len(req.Header.peek(strContentEncoding)) == 0 {
				req.multipartForm, err = readMultipartForm(r, req.multipartFormBoundary, contentLength, defaultMaxInMemoryFileSize)
				if err != nil {
					req.Reset()
				}
				return err
			}
		}
	}

	if contentLength == -2 {
		// identity body has no sense for http requests, since
		// the end of body is determined by connection close.
		// So just ignore request body for requests without
		// 'Content-Length' and 'Transfer-Encoding' headers.

		// refer to https://tools.ietf.org/html/rfc7230#section-3.3.2
		if !req.Header.ignoreBody() {
			req.Header.SetContentLength(0)
		}
		return nil
	}

	bodyBuf := req.bodyBuffer()
	bodyBuf.Reset()
	bodyBuf.B, err = readBodyWithStreaming(r, contentLength, maxBodySize, bodyBuf.B)
	if err != nil {
		if err == ErrBodyTooLarge {
			req.Header.SetContentLength(contentLength)
			req.body = bodyBuf
			req.bodyStream = acquireRequestStream(bodyBuf, r, &req.Header)
			return nil
		}
		if err == errChunkedStream {
			req.body = bodyBuf
			req.bodyStream = acquireRequestStream(bodyBuf, r, &req.Header)
			return nil
		}
		req.Reset()
		return err
	}

	req.body = bodyBuf
	req.bodyStream = acquireRequestStream(bodyBuf, r, &req.Header)
	req.Header.SetContentLength(contentLength)
	return nil
}

// Read reads response (including body) from the given r.
//
// io.EOF is returned if r is closed before reading the first header byte.
func (resp *Response) Read(r *bufio.Reader) error {
	return resp.ReadLimitBody(r, 0)
}

// ReadLimitBody reads response headers from the given r,
// then reads the body using the ReadBody function and limiting the body size.
//
// If resp.SkipBody is true then it skips reading the response body.
//
// If maxBodySize > 0 and the body size exceeds maxBodySize,
// then ErrBodyTooLarge is returned.
//
// io.EOF is returned if r is closed before reading the first header byte.
func (resp *Response) ReadLimitBody(r *bufio.Reader, maxBodySize int) error {
	resp.resetSkipHeader()
	err := resp.Header.Read(r)
	if err != nil {
		return err
	}
	if resp.Header.StatusCode() == StatusContinue {
		// Read the next response according to http://www.w3.org/Protocols/rfc2616/rfc2616-sec8.html .
		if err = resp.Header.Read(r); err != nil {
			return err
		}
	}

	if !resp.mustSkipBody() {
		err = resp.ReadBody(r, maxBodySize)
		if err != nil {
			return err
		}
	}

	// A response without a body can't have trailers.
	if resp.Header.ContentLength() == -1 && !resp.StreamBody && !resp.mustSkipBody() {
		err = resp.Header.ReadTrailer(r)
		if err != nil && err != io.EOF {
			return err
		}
	}
	return nil
}

// ReadBody reads response body from the given r, limiting the body size.
//
// If maxBodySize > 0 and the body size exceeds maxBodySize,
// then ErrBodyTooLarge is returned.
func (resp *Response) ReadBody(r *bufio.Reader, maxBodySize int) (err error) {
	bodyBuf := resp.bodyBuffer()
	bodyBuf.Reset()

	contentLength := resp.Header.ContentLength()
	switch {
	case contentLength >= 0:
		bodyBuf.B, err = readBody(r, contentLength, maxBodySize, bodyBuf.B)
		if err == ErrBodyTooLarge && resp.StreamBody {
			resp.bodyStream = acquireRequestStream(bodyBuf, r, &resp.Header)
			err = nil
		}
	case contentLength == -1:
		if resp.StreamBody {
			resp.bodyStream = acquireRequestStream(bodyBuf, r, &resp.Header)
		} else {
			bodyBuf.B, err = readBodyChunked(r, maxBodySize, bodyBuf.B)
		}
	default:
		if resp.StreamBody {
			resp.bodyStream = acquireRequestStream(bodyBuf, r, &resp.Header)
		} else {
			bodyBuf.B, err = readBodyIdentity(r, maxBodySize, bodyBuf.B)
			resp.Header.SetContentLength(len(bodyBuf.B))
		}
	}
	if err == nil && resp.StreamBody && resp.bodyStream == nil {
		resp.bodyStream = bytes.NewReader(bodyBuf.B)
	}
	return err
}

func (resp *Response) mustSkipBody() bool {
	return resp.SkipBody || resp.Header.mustSkipContentLength()
}

var errRequestHostRequired = errors.New("missing required Host header in request")

// WriteTo writes request to w. It implements io.WriterTo.
func (req *Request) WriteTo(w io.Writer) (int64, error) {
	return writeBufio(req, w)
}

// WriteTo writes response to w. It implements io.WriterTo.
func (resp *Response) WriteTo(w io.Writer) (int64, error) {
	return writeBufio(resp, w)
}

func writeBufio(hw httpWriter, w io.Writer) (int64, error) {
	sw := acquireStatsWriter(w)
	bw := acquireBufioWriter(sw)
	errw := hw.Write(bw)
	errf := bw.Flush()
	releaseBufioWriter(bw)
	n := sw.bytesWritten
	releaseStatsWriter(sw)

	err := errw
	if err == nil {
		err = errf
	}
	return n, err
}

type statsWriter struct {
	w            io.Writer
	bytesWritten int64
}

func (w *statsWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.bytesWritten += int64(n)
	return n, err
}

func (w *statsWriter) WriteString(s string) (int, error) {
	n, err := w.w.Write(s2b(s))
	w.bytesWritten += int64(n)
	return n, err
}

func acquireStatsWriter(w io.Writer) *statsWriter {
	v := statsWriterPool.Get()
	if v == nil {
		return &statsWriter{
			w: w,
		}
	}
	sw := v.(*statsWriter)
	sw.w = w
	return sw
}

func releaseStatsWriter(sw *statsWriter) {
	sw.w = nil
	sw.bytesWritten = 0
	statsWriterPool.Put(sw)
}

var statsWriterPool sync.Pool

func acquireBufioWriter(w io.Writer) *bufio.Writer {
	v := bufioWriterPool.Get()
	if v == nil {
		return bufio.NewWriter(w)
	}
	bw := v.(*bufio.Writer)
	bw.Reset(w)
	return bw
}

func releaseBufioWriter(bw *bufio.Writer) {
	bufioWriterPool.Put(bw)
}

var bufioWriterPool sync.Pool

func (req *Request) onlyMultipartForm() bool {
	return req.multipartForm != nil && (req.body == nil || len(req.body.B) == 0)
}

// Write writes request to w.
//
// Write doesn't flush request to w for performance reasons.
//
// See also WriteTo.
func (req *Request) Write(w *bufio.Writer) error {
	if len(req.Header.Host()) == 0 || req.parsedURI {
		uri := req.URI()
		host := uri.Host()
		if len(req.Header.Host()) == 0 {
			if len(host) == 0 {
				return errRequestHostRequired
			}
			req.Header.SetHostBytes(host)
		} else if !req.UseHostHeader {
			req.Header.SetHostBytes(host)
		}
		req.Header.SetRequestURIBytes(uri.RequestURI())

		if len(uri.username) > 0 {
			// RequestHeader.SetBytesKV only uses RequestHeader.bufKV.key
			// So we are free to use RequestHeader.bufKV.value as a scratch pad for
			// the base64 encoding.
			nl := len(uri.username) + len(uri.password) + 1
			nb := nl + len(strBasicSpace)
			tl := nb + base64.StdEncoding.EncodedLen(nl)
			if tl > cap(req.Header.bufV) {
				req.Header.bufV = make([]byte, 0, tl)
			}
			buf := req.Header.bufV[:0]
			buf = append(buf, uri.username...)
			buf = append(buf, strColon...)
			buf = append(buf, uri.password...)
			buf = append(buf, strBasicSpace...)
			base64.StdEncoding.Encode(buf[nb:tl], buf[:nl])
			req.Header.SetBytesKV(strAuthorization, buf[nl:tl])
		}
	}

	if req.bodyStream != nil {
		return req.writeBodyStream(w)
	}

	body := req.bodyBytes()
	var err error
	if req.onlyMultipartForm() {
		body, err = marshalMultipartForm(req.multipartForm, req.multipartFormBoundary)
		if err != nil {
			return fmt.Errorf("error when marshaling multipart form: %w", err)
		}
		req.Header.SetMultipartFormBoundary(req.multipartFormBoundary)
	}

	hasBody := false
	if len(body) == 0 {
		body = req.postArgs.QueryString()
	}
	if len(body) != 0 || !req.Header.ignoreBody() {
		hasBody = true
		req.Header.SetContentLength(len(body))
	}
	if err = req.Header.Write(w); err != nil {
		return err
	}
	if hasBody {
		_, err = w.Write(body)
	} else if len(body) > 0 {
		if req.secureErrorLogMessage {
			return errors.New("non-zero body for non-POST request")
		}
		return fmt.Errorf("non-zero body for non-POST request. body=%q", body)
	}
	return err
}

// WriteGzip writes response with gzipped body to w.
//
// The method gzips response body and sets 'Content-Encoding: gzip'
// header before writing response to w.
//
// WriteGzip doesn't flush response to w for performance reasons.
func (resp *Response) WriteGzip(w *bufio.Writer) error {
	return resp.WriteGzipLevel(w, CompressDefaultCompression)
}

// WriteGzipLevel writes response with gzipped body to w.
//
// Level is the desired compression level:
//
//   - CompressNoCompression
//   - CompressBestSpeed
//   - CompressBestCompression
//   - CompressDefaultCompression
//   - CompressHuffmanOnly
//
// The method gzips response body and sets 'Content-Encoding: gzip'
// header before writing response to w.
//
// WriteGzipLevel doesn't flush response to w for performance reasons.
func (resp *Response) WriteGzipLevel(w *bufio.Writer, level int) error {
	resp.gzipBody(level)
	return resp.Write(w)
}

// WriteDeflate writes response with deflated body to w.
//
// The method deflates response body and sets 'Content-Encoding: deflate'
// header before writing response to w.
//
// WriteDeflate doesn't flush response to w for performance reasons.
func (resp *Response) WriteDeflate(w *bufio.Writer) error {
	return resp.WriteDeflateLevel(w, CompressDefaultCompression)
}

// WriteDeflateLevel writes response with deflated body to w.
//
// Level is the desired compression level:
//
//   - CompressNoCompression
//   - CompressBestSpeed
//   - CompressBestCompression
//   - CompressDefaultCompression
//   - CompressHuffmanOnly
//
// The method deflates response body and sets 'Content-Encoding: deflate'
// header before writing response to w.
//
// WriteDeflateLevel doesn't flush response to w for performance reasons.
func (resp *Response) WriteDeflateLevel(w *bufio.Writer, level int) error {
	resp.deflateBody(level)
	return resp.Write(w)
}

func (resp *Response) brotliBody(level int) {
	if len(resp.Header.ContentEncoding()) > 0 {
		// It looks like the body is already compressed.
		// Do not compress it again.
		return
	}

	if !resp.Header.isCompressibleContentType() {
		// The content-type cannot be compressed.
		return
	}

	if resp.bodyStream != nil {
		// Reset Content-Length to -1, since it is impossible
		// to determine body size beforehand of streamed compression.
		// For https://github.com/valyala/fasthttp/issues/176 .
		resp.Header.SetContentLength(-1)

		// Do not care about memory allocations here, since brotli is slow
		// and allocates a lot of memory by itself.
		bs := resp.bodyStream
		resp.bodyStream = NewStreamReader(func(sw *bufio.Writer) {
			zw := acquireStacklessBrotliWriter(sw, level)
			fw := &flushWriter{
				wf: zw,
				bw: sw,
			}
			_, wErr := copyZeroAlloc(fw, bs)
			releaseStacklessBrotliWriter(zw, level)
			switch v := bs.(type) {
			case io.Closer:
				v.Close()
			case ReadCloserWithError:
				v.CloseWithError(wErr) //nolint:errcheck
			}
		})
	} else {
		bodyBytes := resp.bodyBytes()
		if len(bodyBytes) < minCompressLen {
			// There is no sense in spending CPU time on small body compression,
			// since there is a very high probability that the compressed
			// body size will be bigger than the original body size.
			return
		}
		w := responseBodyPool.Get()
		w.B = AppendBrotliBytesLevel(w.B, bodyBytes, level)

		// Hack: swap resp.body with w.
		if resp.body != nil {
			responseBodyPool.Put(resp.body)
		}
		resp.body = w
		resp.bodyRaw = nil
	}
	resp.Header.SetContentEncodingBytes(strBr)
	resp.Header.addVaryBytes(strAcceptEncoding)
}

func (resp *Response) gzipBody(level int) {
	if len(resp.Header.ContentEncoding()) > 0 {
		// It looks like the body is already compressed.
		// Do not compress it again.
		return
	}

	if !resp.Header.isCompressibleContentType() {
		// The content-type cannot be compressed.
		return
	}

	if resp.bodyStream != nil {
		// Reset Content-Length to -1, since it is impossible
		// to determine body size beforehand of streamed compression.
		// For https://github.com/valyala/fasthttp/issues/176 .
		resp.Header.SetContentLength(-1)

		// Do not care about memory allocations here, since gzip is slow
		// and allocates a lot of memory by itself.
		bs := resp.bodyStream
		resp.bodyStream = NewStreamReader(func(sw *bufio.Writer) {
			zw := acquireStacklessGzipWriter(sw, level)
			fw := &flushWriter{
				wf: zw,
				bw: sw,
			}
			_, wErr := copyZeroAlloc(fw, bs)
			releaseStacklessGzipWriter(zw, level)
			switch v := bs.(type) {
			case io.Closer:
				v.Close()
			case ReadCloserWithError:
				v.CloseWithError(wErr) //nolint:errcheck
			}
		})
	} else {
		bodyBytes := resp.bodyBytes()
		if len(bodyBytes) < minCompressLen {
			// There is no sense in spending CPU time on small body compression,
			// since there is a very high probability that the compressed
			// body size will be bigger than the original body size.
			return
		}
		w := responseBodyPool.Get()
		w.B = AppendGzipBytesLevel(w.B, bodyBytes, level)

		// Hack: swap resp.body with w.
		if resp.body != nil {
			responseBodyPool.Put(resp.body)
		}
		resp.body = w
		resp.bodyRaw = nil
	}
	resp.Header.SetContentEncodingBytes(strGzip)
	resp.Header.addVaryBytes(strAcceptEncoding)
}

func (resp *Response) deflateBody(level int) {
	if len(resp.Header.ContentEncoding()) > 0 {
		// It looks like the body is already compressed.
		// Do not compress it again.
		return
	}

	if !resp.Header.isCompressibleContentType() {
		// The content-type cannot be compressed.
		return
	}

	if resp.bodyStream != nil {
		// Reset Content-Length to -1, since it is impossible
		// to determine body size beforehand of streamed compression.
		// For https://github.com/valyala/fasthttp/issues/176 .
		resp.Header.SetContentLength(-1)

		// Do not care about memory allocations here, since flate is slow
		// and allocates a lot of memory by itself.
		bs := resp.bodyStream
		resp.bodyStream = NewStreamReader(func(sw *bufio.Writer) {
			zw := acquireStacklessDeflateWriter(sw, level)
			fw := &flushWriter{
				wf: zw,
				bw: sw,
			}
			_, wErr := copyZeroAlloc(fw, bs)
			releaseStacklessDeflateWriter(zw, level)
			switch v := bs.(type) {
			case io.Closer:
				v.Close()
			case ReadCloserWithError:
				v.CloseWithError(wErr) //nolint:errcheck
			}
		})
	} else {
		bodyBytes := resp.bodyBytes()
		if len(bodyBytes) < minCompressLen {
			// There is no sense in spending CPU time on small body compression,
			// since there is a very high probability that the compressed
			// body size will be bigger than the original body size.
			return
		}
		w := responseBodyPool.Get()
		w.B = AppendDeflateBytesLevel(w.B, bodyBytes, level)

		// Hack: swap resp.body with w.
		if resp.body != nil {
			responseBodyPool.Put(resp.body)
		}
		resp.body = w
		resp.bodyRaw = nil
	}
	resp.Header.SetContentEncodingBytes(strDeflate)
	resp.Header.addVaryBytes(strAcceptEncoding)
}

func (resp *Response) zstdBody(level int) {
	if len(resp.Header.ContentEncoding()) > 0 {
		return
	}

	if !resp.Header.isCompressibleContentType() {
		return
	}

	if resp.bodyStream != nil {
		// Reset Content-Length to -1, since it is impossible
		// to determine body size beforehand of streamed compression.
		// For
		resp.Header.SetContentLength(-1)

		// Do not care about memory allocations here, since flate is slow
		// and allocates a lot of memory by itself.
		bs := resp.bodyStream
		resp.bodyStream = NewStreamReader(func(sw *bufio.Writer) {
			zw := acquireStacklessZstdWriter(sw, level)
			fw := &flushWriter{
				wf: zw,
				bw: sw,
			}
			_, wErr := copyZeroAlloc(fw, bs)
			releaseStacklessZstdWriter(zw, level)
			switch v := bs.(type) {
			case io.Closer:
				v.Close()
			case ReadCloserWithError:
				v.CloseWithError(wErr) //nolint:errcheck
			}
		})
	} else {
		bodyBytes := resp.bodyBytes()
		if len(bodyBytes) < minCompressLen {
			return
		}
		w := responseBodyPool.Get()
		w.B = AppendZstdBytesLevel(w.B, bodyBytes, level)

		if resp.body != nil {
			responseBodyPool.Put(resp.body)
		}
		resp.body = w
		resp.bodyRaw = nil
	}
	resp.Header.SetContentEncodingBytes(strZstd)
	resp.Header.addVaryBytes(strAcceptEncoding)
}

// Bodies with sizes smaller than minCompressLen aren't compressed at all.
const minCompressLen = 200

type writeFlusher interface {
	io.Writer
	Flush() error
}

type flushWriter struct {
	wf writeFlusher
	bw *bufio.Writer
}

func (w *flushWriter) Write(p []byte) (int, error) {
	n, err := w.wf.Write(p)
	if err != nil {
		return 0, err
	}
	if err = w.wf.Flush(); err != nil {
		return 0, err
	}
	if err = w.bw.Flush(); err != nil {
		return 0, err
	}
	return n, nil
}

func (w *flushWriter) WriteString(s string) (int, error) {
	return w.Write(s2b(s))
}

// Write writes response to w.
//
// Write doesn't flush response to w for performance reasons.
//
// See also WriteTo.
func (resp *Response) Write(w *bufio.Writer) error {
	sendBody := !resp.mustSkipBody()

	if resp.bodyStream != nil {
		return resp.writeBodyStream(w, sendBody)
	}

	body := resp.bodyBytes()
	bodyLen := len(body)
	if sendBody || bodyLen > 0 {
		resp.Header.SetContentLength(bodyLen)
	}
	if err := resp.Header.Write(w); err != nil {
		return err
	}
	if sendBody {
		if _, err := w.Write(body); err != nil {
			return err
		}
	}
	return nil
}

func (req *Request) writeBodyStream(w *bufio.Writer) error {
	var err error

	contentLength := req.Header.ContentLength()
	if contentLength < 0 {
		lrSize := limitedReaderSize(req.bodyStream)
		if lrSize >= 0 {
			contentLength = int(lrSize)
			if int64(contentLength) != lrSize {
				contentLength = -1
			}
			if contentLength >= 0 {
				req.Header.SetContentLength(contentLength)
			}
		}
	}
	if contentLength >= 0 {
		if err = req.Header.Write(w); err == nil {
			err = writeBodyFixedSize(w, req.bodyStream, int64(contentLength))
		}
	} else {
		req.Header.SetContentLength(-1)
		err = req.Header.Write(w)
		if err == nil {
			err = writeBodyChunked(w, req.bodyStream)
		}
		if err == nil {
			err = req.Header.writeTrailer(w)
		}
	}
	errc := req.closeBodyStream()
	if err == nil {
		err = errc
	}
	return err
}

// ErrBodyStreamWritePanic is returned when panic happens during writing body stream.
type ErrBodyStreamWritePanic struct {
	error
}

func (resp *Response) writeBodyStream(w *bufio.Writer, sendBody bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &ErrBodyStreamWritePanic{
				error: fmt.Errorf("panic while writing body stream: %+v", r),
			}
		}
	}()

	contentLength := resp.Header.ContentLength()
	if contentLength < 0 {
		lrSize := limitedReaderSize(resp.bodyStream)
		if lrSize >= 0 {
			contentLength = int(lrSize)
			if int64(contentLength) != lrSize {
				contentLength = -1
			}
			if contentLength >= 0 {
				resp.Header.SetContentLength(contentLength)
			}
		}
	}
	if contentLength >= 0 {
		if err = resp.Header.Write(w); err == nil {
			if resp.ImmediateHeaderFlush {
				err = w.Flush()
			}
			if err == nil && sendBody {
				err = writeBodyFixedSize(w, resp.bodyStream, int64(contentLength))
			}
		}
	} else {
		resp.Header.SetContentLength(-1)
		if err = resp.Header.Write(w); err == nil {
			if resp.ImmediateHeaderFlush {
				err = w.Flush()
			}
			if err == nil && sendBody {
				err = writeBodyChunked(w, resp.bodyStream)
			}
			if err == nil {
				err = resp.Header.writeTrailer(w)
			}
		}
	}
	errc := resp.closeBodyStream(err)
	if err == nil {
		err = errc
	}
	return err
}

func (req *Request) closeBodyStream() error {
	if req.bodyStream == nil {
		return nil
	}
	var err error
	if bsc, ok := req.bodyStream.(io.Closer); ok {
		err = bsc.Close()
	}
	if rs, ok := req.bodyStream.(*requestStream); ok {
		releaseRequestStream(rs)
	}
	req.bodyStream = nil
	return err
}

func (resp *Response) closeBodyStream(wErr error) error {
	if resp.bodyStream == nil {
		return nil
	}
	var err error
	if bsc, ok := resp.bodyStream.(io.Closer); ok {
		err = bsc.Close()
	}
	if bsc, ok := resp.bodyStream.(ReadCloserWithError); ok {
		err = bsc.CloseWithError(wErr)
	}
	if bsr, ok := resp.bodyStream.(*requestStream); ok {
		releaseRequestStream(bsr)
	}
	resp.bodyStream = nil
	return err
}

// String returns request representation.
//
// Returns error message instead of request representation on error.
//
// Use Write instead of String for performance-critical code.
func (req *Request) String() string {
	return getHTTPString(req)
}

// String returns response representation.
//
// Returns error message instead of response representation on error.
//
// Use Write instead of String for performance-critical code.
func (resp *Response) String() string {
	return getHTTPString(resp)
}

// SetUserValue stores the given value (arbitrary object)
// under the given key in Request.
//
// The value stored in Request may be obtained by UserValue*.
//
// This functionality may be useful for passing arbitrary values between
// functions involved in request processing.
//
// All the values are removed from Request after returning from the top
// RequestHandler. Additionally, Close method is called on each value
// implementing io.Closer before removing the value from Request.
func (req *Request) SetUserValue(key, value any) {
	req.userValues.Set(key, value)
}

// SetUserValueBytes stores the given value (arbitrary object)
// under the given key in Request.
//
// The value stored in Request may be obtained by UserValue*.
//
// This functionality may be useful for passing arbitrary values between
// functions involved in request processing.
//
// All the values stored in Request are deleted after returning from RequestHandler.
func (req *Request) SetUserValueBytes(key []byte, value any) {
	req.userValues.SetBytes(key, value)
}

// UserValue returns the value stored via SetUserValue* under the given key.
func (req *Request) UserValue(key any) any {
	return req.userValues.Get(key)
}

// UserValueBytes returns the value stored via SetUserValue*
// under the given key.
func (req *Request) UserValueBytes(key []byte) any {
	return req.userValues.GetBytes(key)
}

// VisitUserValues calls visitor for each existing userValue with a key that is a string or []byte.
//
// visitor must not retain references to key and value after returning.
// Make key and/or value copies if you need storing them after returning.
func (req *Request) VisitUserValues(visitor func([]byte, any)) {
	for i, n := 0, len(req.userValues); i < n; i++ {
		kv := &req.userValues[i]
		if _, ok := kv.key.(string); ok {
			visitor(s2b(kv.key.(string)), kv.value)
		}
	}
}

// VisitUserValuesAll calls visitor for each existing userValue.
//
// visitor must not retain references to key and value after returning.
// Make key and/or value copies if you need storing them after returning.
func (req *Request) VisitUserValuesAll(visitor func(any, any)) {
	for i, n := 0, len(req.userValues); i < n; i++ {
		kv := &req.userValues[i]
		visitor(kv.key, kv.value)
	}
}

// ResetUserValues allows to reset user values from Request Context.
func (req *Request) ResetUserValues() {
	req.userValues.Reset()
}

// RemoveUserValue removes the given key and the value under it in Request.
func (req *Request) RemoveUserValue(key any) {
	req.userValues.Remove(key)
}

// RemoveUserValueBytes removes the given key and the value under it in Request.
func (req *Request) RemoveUserValueBytes(key []byte) {
	req.userValues.RemoveBytes(key)
}

func getHTTPString(hw httpWriter) string {
	w := bytebufferpool.Get()
	defer bytebufferpool.Put(w)

	bw := bufio.NewWriter(w)
	if err := hw.Write(bw); err != nil {
		return err.Error()
	}
	if err := bw.Flush(); err != nil {
		return err.Error()
	}
	s := string(w.B)
	return s
}

type httpWriter interface {
	Write(w *bufio.Writer) error
}

func writeBodyChunked(w *bufio.Writer, r io.Reader) error {
	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)

	var err error
	var n int
	for {
		n, err = r.Read(buf)
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				if err = writeChunk(w, buf[:0]); err != nil {
					break
				}
				err = nil
			}
			break
		}
		if err = writeChunk(w, buf[:n]); err != nil {
			break
		}
	}

	copyBufPool.Put(vbuf)
	return err
}

func limitedReaderSize(r io.Reader) int64 {
	lr, ok := r.(*io.LimitedReader)
	if !ok {
		return -1
	}
	return lr.N
}

func writeBodyFixedSize(w *bufio.Writer, r io.Reader, size int64) error {
	if size > maxSmallFileSize {
		earlyFlush := false
		switch r := r.(type) {
		case *os.File:
			earlyFlush = true
		case *io.LimitedReader:
			_, earlyFlush = r.R.(*os.File)
		}
		if earlyFlush {
			// w buffer must be empty for triggering
			// sendfile path in bufio.Writer.ReadFrom.
			if err := w.Flush(); err != nil {
				return err
			}
		}
	}

	n, err := copyZeroAlloc(w, r)

	if n != size && err == nil {
		err = fmt.Errorf("copied %d bytes from body stream instead of %d bytes", n, size)
	}
	return err
}

// copyZeroAlloc optimizes io.Copy by calling ReadFrom or WriteTo only when
// copying between os.File and net.TCPConn. If the reader has a WriteTo
// method, it uses WriteTo for copying; if the writer has a ReadFrom method,
// it uses ReadFrom for copying. If neither method is available, it gets a
// buffer from sync.Pool to perform the copy.
//
// io.CopyBuffer always uses the WriterTo or ReadFrom interface if it's
// available. however, os.File and net.TCPConn unfortunately have a
// fallback in their WriterTo that calls io.Copy if sendfile isn't possible.
//
// See issue: https://github.com/valyala/fasthttp/issues/1889
//
// sendfile can only be triggered when copying between os.File and net.TCPConn.
// Since the function confirming zero-copy is a private function, we use
// ReadFrom only in this specific scenario. For all other cases, we prioritize
// using our own copyBuffer method.
//
// o: our copyBuffer
// r: readFrom
// w: writeTo
//
// write\read *File  *TCPConn  writeTo  other
// *File        o       r         w       o
// *TCPConn    w,r      o         w       o
// readFrom     r       r         w       r
// other        o       o         w       o
//
//nolint:dupword
func copyZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	var readerIsFile, readerIsConn bool

	switch r := r.(type) {
	case *os.File:
		readerIsFile = true
	case *net.TCPConn:
		readerIsConn = true
	case io.WriterTo:
		return r.WriteTo(w)
	}

	switch w := w.(type) {
	case *os.File:
		if readerIsConn {
			return w.ReadFrom(r)
		}
	case *net.TCPConn:
		if readerIsFile {
			// net.WriteTo requires go1.22 or later
			// Benchmark tests show that on Windows, WriteTo performs
			// significantly better than ReadFrom. On Linux, however,
			// ReadFrom slightly outperforms WriteTo. When possible,
			// copyZeroAlloc aims to perform  better than or as well
			// as io.Copy, so we use WriteTo whenever possible for
			// optimal performance.
			if rt, ok := r.(io.WriterTo); ok {
				return rt.WriteTo(w)
			}
			return w.ReadFrom(r)
		}
	case io.ReaderFrom:
		return w.ReadFrom(r)
	}

	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := copyBuffer(w, r, buf)
	copyBufPool.Put(vbuf)
	return n, err
}

// copyBuffer is rewritten from io.copyBuffer. We do not check if src has a
// WriteTo method, if dst has a ReadFrom method, or if buf is empty.
func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

var copyBufPool = sync.Pool{
	New: func() any {
		return make([]byte, 4096)
	},
}

func writeChunk(w *bufio.Writer, b []byte) error {
	n := len(b)
	if err := writeHexInt(w, n); err != nil {
		return err
	}
	if _, err := w.Write(strCRLF); err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	// If is end chunk, write CRLF after writing trailer
	if n > 0 {
		if _, err := w.Write(strCRLF); err != nil {
			return err
		}
	}
	return w.Flush()
}

// ErrBodyTooLarge is returned if either request or response body exceeds
// the given limit.
var ErrBodyTooLarge = errors.New("body size exceeds the given limit")

func readBody(r *bufio.Reader, contentLength, maxBodySize int, dst []byte) ([]byte, error) {
	if maxBodySize > 0 && contentLength > maxBodySize {
		return dst, ErrBodyTooLarge
	}
	return appendBodyFixedSize(r, dst, contentLength)
}

var errChunkedStream = errors.New("chunked stream")

func readBodyWithStreaming(r *bufio.Reader, contentLength, maxBodySize int, dst []byte) (b []byte, err error) {
	if contentLength == -1 {
		// handled in requestStream.Read()
		return b, errChunkedStream
	}

	dst = dst[:0]

	readN := maxBodySize
	if readN > contentLength {
		readN = contentLength
	}
	if readN > 8*1024 {
		readN = 8 * 1024
	}

	// A fixed-length pre-read function should be used here; otherwise,
	// it may read content beyond the request body into areas outside
	// the br buffer. This could affect the handling of the next request
	// in the br buffer, if there is one. The original two branches can
	// be handled with this single branch. by the way,
	// fix issue: https://github.com/valyala/fasthttp/issues/1816
	b, err = appendBodyFixedSize(r, dst, readN)
	if err != nil {
		return b, err
	}
	if contentLength > maxBodySize {
		return b, ErrBodyTooLarge
	}
	return b, nil
}

func readBodyIdentity(r *bufio.Reader, maxBodySize int, dst []byte) ([]byte, error) {
	dst = dst[:cap(dst)]
	if len(dst) == 0 {
		dst = make([]byte, 1024)
	}
	offset := 0
	for {
		nn, err := r.Read(dst[offset:])
		if nn <= 0 {
			switch {
			case errors.Is(err, io.EOF):
				return dst[:offset], nil
			case err != nil:
				return dst[:offset], err
			default:
				return dst[:offset], fmt.Errorf("bufio.Read() returned (%d, nil)", nn)
			}
		}
		offset += nn
		if maxBodySize > 0 && offset > maxBodySize {
			return dst[:offset], ErrBodyTooLarge
		}
		if len(dst) == offset {
			n := roundUpForSliceCap(2 * offset)
			if maxBodySize > 0 && n > maxBodySize {
				n = maxBodySize + 1
			}
			b := make([]byte, n)
			copy(b, dst)
			dst = b
		}
	}
}

func appendBodyFixedSize(r *bufio.Reader, dst []byte, n int) ([]byte, error) {
	if n == 0 {
		return dst, nil
	}

	offset := len(dst)
	dstLen := offset + n
	if cap(dst) < dstLen {
		b := make([]byte, roundUpForSliceCap(dstLen))
		copy(b, dst)
		dst = b
	}
	dst = dst[:dstLen]

	for {
		nn, err := r.Read(dst[offset:])
		if nn <= 0 {
			switch {
			case errors.Is(err, io.EOF):
				return dst[:offset], io.ErrUnexpectedEOF
			case err != nil:
				return dst[:offset], err
			default:
				return dst[:offset], fmt.Errorf("bufio.Read() returned (%d, nil)", nn)
			}
		}
		offset += nn
		if offset == dstLen {
			return dst, nil
		}
	}
}

// ErrBrokenChunk is returned when server receives a broken chunked body (Transfer-Encoding: chunked).
type ErrBrokenChunk struct {
	error
}

func readBodyChunked(r *bufio.Reader, maxBodySize int, dst []byte) ([]byte, error) {
	if len(dst) > 0 {
		// data integrity might be in danger. No idea what we received,
		// but nothing we should write to.
		panic("BUG: expected zero-length buffer")
	}

	strCRLFLen := len(strCRLF)
	for {
		chunkSize, err := parseChunkSize(r)
		if err != nil {
			return dst, err
		}
		if chunkSize == 0 {
			return dst, err
		}
		if maxBodySize > 0 && len(dst)+chunkSize > maxBodySize {
			return dst, ErrBodyTooLarge
		}
		dst, err = appendBodyFixedSize(r, dst, chunkSize+strCRLFLen)
		if err != nil {
			return dst, err
		}
		if !bytes.Equal(dst[len(dst)-strCRLFLen:], strCRLF) {
			return dst, ErrBrokenChunk{
				error: errors.New("cannot find crlf at the end of chunk"),
			}
		}
		dst = dst[:len(dst)-strCRLFLen]
	}
}

func parseChunkSize(r *bufio.Reader) (int, error) {
	n, err := readHexInt(r)
	if err != nil {
		return -1, err
	}
	for {
		c, err := r.ReadByte()
		if err != nil {
			return -1, ErrBrokenChunk{
				error: fmt.Errorf("cannot read '\\r' char at the end of chunk size: %w", err),
			}
		}
		// Skip chunk extension after chunk size.
		// Add support later if anyone needs it.
		if c != '\r' {
			// Security: Don't allow newlines in chunk extensions.
			// This can lead to request smuggling issues with some reverse proxies.
			if c == '\n' {
				return -1, ErrBrokenChunk{
					error: errors.New("invalid character '\\n' after chunk size"),
				}
			}
			continue
		}
		if err := r.UnreadByte(); err != nil {
			return -1, ErrBrokenChunk{
				error: fmt.Errorf("cannot unread '\\r' char at the end of chunk size: %w", err),
			}
		}
		break
	}
	err = readCrLf(r)
	if err != nil {
		return -1, err
	}
	return n, nil
}

func readCrLf(r *bufio.Reader) error {
	for _, exp := range []byte{'\r', '\n'} {
		c, err := r.ReadByte()
		if err != nil {
			return ErrBrokenChunk{
				error: fmt.Errorf("cannot read %q char at the end of chunk size: %w", exp, err),
			}
		}
		if c != exp {
			return ErrBrokenChunk{
				error: fmt.Errorf("unexpected char %q at the end of chunk size. Expected %q", c, exp),
			}
		}
	}
	return nil
}

// SetTimeout sets timeout for the request.
//
// The following code:
//
//	req.SetTimeout(t)
//	c.Do(&req, &resp)
//
// is equivalent to
//
//	c.DoTimeout(&req, &resp, t)
func (req *Request) SetTimeout(t time.Duration) {
	req.timeout = t
}
