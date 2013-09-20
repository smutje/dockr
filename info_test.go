package dockr

import (
  "testing"
)

func TestVersion(t *testing.T){
  client := newTestClient()
  s, err := client.Version()
  if err != nil {
    t.Fatal(err)
  }
  t.Log(s)
}
