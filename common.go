package dockr

import (
  "net"
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
}

func (err *UnexpectedHTTPStatus) Error() string {
  return fmt.Sprintf("Unexpected HTTP status %d, expected %v",err.Actual, err.Expected)
}

func validateId(id string) error {
  if !ID_REGEXP.MatchString(id) {
    return InvalidId(id)
  }
  return nil
}

func expectHTTPStatus( actual int, expected ...int) error {
  for _, e := range expected {
    if e == actual {
      return nil
    }
  }
  return &UnexpectedHTTPStatus{ actual, expected }
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

