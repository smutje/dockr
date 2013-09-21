package dockr

import (
  "time"
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
  ch := BuildStatusScanner(rc)
  select{
    case <-ch :
    case <-time.After(5 * time.Second) :
      t.Fatal("Timed out")
  }
  rc.Close()
}
