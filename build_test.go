package dockr

import (
  "os"
  "io"
 // "bufio"
  "testing"
)

var (
  DOCKER_FILE = `FROM ubuntu:precise
RUN echo pong > /ping`
)

func TestBuild(t *testing.T) {
  client := newTestClient()
  df,_   := SimpleDockerFile(DOCKER_FILE)
  rc, err := client.Build(&BuildRequest{Body: df})
  if err != nil {
    t.Fatal(err)
  }
  io.Copy(os.Stderr, rc)
  /*rd := bufio.NewReader(rc)
  for {
    str, err := rd.ReadString('\n')
    if err != nil {
      t.Fatal(err)
    }
    t.Log(str)
  }*/
}
