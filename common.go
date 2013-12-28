package dockr

import (
  "net"
  "net/http"
  "io"
  "fmt"
  "regexp"
)

var (
  ID_REGEXP = regexp.MustCompile("\\A[0-9a-fA-F]{12,70}\\z")
  NOT_FOUND = fmt.Errorf("Not found")
)

type InvalidId string
func (err InvalidId) Error() string {
  return fmt.Sprintf("Invalid id: \"%20s...\"", string(err))
}

type UnexpectedHTTPStatus struct{
  Actual   int
  Expected []int
  Message  string
}

func (err *UnexpectedHTTPStatus) Error() string {
  return fmt.Sprintf("Unexpected HTTP status %d, expected %v. Message: %v",err.Actual, err.Expected, err.Message)
}

func validateId(id string) error {
  if !ID_REGEXP.MatchString(id) {
    return InvalidId(id)
  }
  return nil
}

func expectHTTPStatus( res *http.Response, expected ...int) error {
  for _, e := range expected {
    if e == res.StatusCode {
      return nil
    }
  }
  var buf []byte
  if res.ContentLength < 0 || res.ContentLength > 1024{
    buf = make([]byte,1024)
  }else{
    buf = make([]byte,res.ContentLength)
  }
  n,_ := io.ReadFull(res.Body, buf)
  // discard all the f***ing errors
  return &UnexpectedHTTPStatus{ res.StatusCode, expected, string(buf[1:n]) }
}

type hijackReadWriteCloser struct {
  con net.Conn
  buf io.Reader
}
func (h *hijackReadWriteCloser) Read(p []byte) (int, error){
  return h.buf.Read(p)
}
func (h *hijackReadWriteCloser) Write(p []byte) (int, error){
  return h.con.Write(p)
}
func (h *hijackReadWriteCloser) Close() error{
  return h.con.Close()
}

