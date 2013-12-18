package dockr

import (
  "fmt"
  "net"
  "net/http"
  "io"
  "bufio"
  "syscall"
  "time"
  "runtime/debug"
)

type testResponseWriter struct {
  out net.Conn
  headerWritten bool
  header http.Header
}

func (w *testResponseWriter) Header() http.Header {
  return w.header
}
func (w *testResponseWriter) WriteHeader(status int) {
  if !w.headerWritten {
    io.WriteString(w.out, fmt.Sprintf("HTTP/1.1 %d %s\n", status, http.StatusText(status)))
    w.header.Write(w.out)
    io.WriteString(w.out,"\n\n")
    w.headerWritten = true
  }
}
func (w *testResponseWriter) Write(b []byte) (int, error) {
  if !w.headerWritten {
    w.WriteHeader(http.StatusOK)
  }
  return w.out.Write(b)
}

func (w *testResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
  return w.out, bufio.NewReadWriter(bufio.NewReader(w.out),bufio.NewWriter(w.out)), nil
}

func newTestResponseWriter(out net.Conn) *testResponseWriter{
  return &testResponseWriter{out, false, make(http.Header)}
}

type testServer struct {
  Handler http.Handler
}

func (t *testServer) Connect() (net.Conn, error){
  a, b := net.Pipe()
  go func(){
    rd := bufio.NewReader(b)
    req, _ := http.ReadRequest(rd)
    t.Handler.ServeHTTP(newTestResponseWriter(b), req)
  }()
  return a, nil
}

func (t *testServer) String() string {
  return "TestServer"
}

type debugNetConn struct {
  inner net.Conn
}

func (c *debugNetConn) Read(b []byte) (n int, err error){
  return c.inner.Read(b)
}
func (c *debugNetConn) Write(b []byte) (n int, err error){
  return c.inner.Write(b)
}
func (c *debugNetConn) Close() error{
  debug.PrintStack()
  return c.inner.Close()
}
func (c *debugNetConn) LocalAddr() net.Addr{
  return c.inner.LocalAddr()
}
func (c *debugNetConn) RemoteAddr() net.Addr{
  return c.inner.RemoteAddr()
}
func (c *debugNetConn) SetDeadline(t time.Time) error{
  return c.inner.SetDeadline(t)
}
func (c *debugNetConn) SetReadDeadline(t time.Time) error{
  return c.inner.SetReadDeadline(t)
}
func (c *debugNetConn) SetWriteDeadline(t time.Time) error{
  return c.inner.SetWriteDeadline(t)
}

type debugConnector struct {
  inner Connector
}
func (c *debugConnector) String() string {
  return c.inner.String()
}
func (c *debugConnector) Connect() (con2 net.Conn,err error) {
  con, err := c.inner.Connect()
  if err != nil {
    return
  }
  con2 = &debugNetConn{con}
  return
}

func newTestClient() *Client {
  h, ok := syscall.Getenv("DOCKR_HOST")
  if !ok {
    h = "tcp://localhost:14243"
  }
  c := NewClientFromConnector(MustDefaultConnector(h))
  return c
}
