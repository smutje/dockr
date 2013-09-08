package dockr

import (
  "fmt"
  "net"
  "net/http"
  "io"
  "bufio"
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


