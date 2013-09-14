package dockr

import (
  "testing"
  "bufio"
  "time"
)

func TestCreateContainerWithoutAnything(t *testing.T){
  client := newTestClient()
  _, err := client.CreateContainer(&CreateContainerRequest{})
  if err == nil {
    t.Fatal(err)
  }
}

func TestCreateContainer(t *testing.T){
  client := newTestClient()
  res, err := client.CreateContainer(&CreateContainerRequest{
    Cmd:[]string{"uname","-a"},
    Image:"ubuntu:precise",
  })
  if err != nil {
    t.Fatal(err)
  }
  if res == nil {
    t.Fatal("Repsonse was nil")
  }
  err = client.DeleteContainer(res.Id)
  if err != nil {
    t.Fatal(err)
  }
}

func TestStartStopContainer(t *testing.T){
  client := newTestClient()
  res, err := client.CreateContainer(&CreateContainerRequest{
    Cmd:[]string{"echo","pong"},
    Image:"ubuntu:precise",
  })
  if err != nil {
    t.Fatal(err)
  }
  if res == nil {
    t.Fatal("Repsonse was nil")
  }
  err = client.StartContainer(res.Id,&StartContainerRequest{})
  if err != nil {
    t.Fatal(err)
  }
  err = client.StopContainer(res.Id,&StopContainerRequest{Timeout: 3})
  if err != nil {
    t.Fatal(err)
  }
  err = client.DeleteContainer(res.Id)
  if err != nil {
    t.Fatal(err)
  }
}

func TestStartAttachStopContainer(t *testing.T){
  client := newTestClient()
  res, err := client.CreateContainer(&CreateContainerRequest{
    Cmd:[]string{"echo","pong"},
    Image:"ubuntu:precise",
  })
  if err != nil {
    t.Fatal(err)
  }
  if res == nil {
    t.Fatal("Repsonse was nil")
  }
  ch := make(chan error)
  rwc, err := client.AttachContainer(res.Id,&AttachContainerRequest{Logs:true, Stream:true, Stdout:true})
  if err != nil {
    t.Fatal( err )
  }
  rd := bufio.NewReader(rwc)
  go func(){
    str, err := rd.ReadString('\n')
    if err != nil {
      ch <- err
      return
    }
    if str != "pong\n" {
      t.Fatalf("Expected 'pong', got '%s'",str)
    }
    ch <- rwc.Close()
  }()
  err = client.StartContainer(res.Id,&StartContainerRequest{})
  if err != nil {
    t.Fatal(err)
  }
  err = client.StopContainer(res.Id,&StopContainerRequest{Timeout: 3})
  if err != nil {
    t.Fatal(err)
  }
  select {
  case err = <-ch :
    if err != nil {
      t.Fatal(err)
    }
  case <-time.After(3 * time.Second) :
    t.Fatal("Response timed out")
  }
  err = client.DeleteContainer(res.Id)
  if err != nil {
    t.Fatal(err)
  }
}

