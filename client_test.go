package dockr

import (
  "testing"
  "io"
  "strings"
  "net/http"
)

func TestClientConnection(t *testing.T) {
  s := testServer{
    http.HandlerFunc(func(res http.ResponseWriter, req *http.Request){
      if req.Method != "GET" {
        t.Fatalf("Expected method GET, not %s", req.Method)
      }
      if req.URL.Path != "/info" {
        t.Fatalf("Expected path /info, not %s", req.URL.Path)
      }
      res.Header().Add("Content-Type","application/json")
      io.WriteString(res, "{}")
    }),
  }
  client := NewClientFromConnector(&s)
  req,_  := http.NewRequest("GET","/info", nil)
  res, _, err := client.do(req)
  if err != nil {
    t.Fatal(err)
  }
  if res == nil {
    t.Fatal("Response is nil")
  }
}

func TestClientHijackedConnection(t *testing.T) {
  p := testServer{
    http.HandlerFunc(func(res http.ResponseWriter, req *http.Request){
      if req.Method != "GET" {
        t.Fatalf("Expected method GET, not %s", req.Method)
      }
      if req.URL.Path != "/upcase" {
        t.Fatalf("Expected path /upcase, not %s", req.URL.Path)
      }
      res.Header().Add("Content-Type","text/plain")
      res.WriteHeader(200)
      hj, ok := res.(http.Hijacker)
      if !ok {
        t.Fatalf("Not hijackable!")
      }
      conn, rw, _ := hj.Hijack()
      defer conn.Close()
      str, _ := rw.ReadString('\n')
      rw.WriteString(strings.ToUpper(str))
      rw.Flush()
    }),
  }
  client := NewClientFromConnector(&p)
  req,_  := http.NewRequest("GET","/upcase", nil)
  res, con, err := client.do(req)
  if err != nil {
    t.Fatal(err)
  }
  if res == nil {
    t.Fatal("Response is nil")
  }
  o, i := con.Hijack()
  i.ReadString('\n')
  io.WriteString(o, "lol\n")
  rs, _ := i.ReadString('\n')
  if rs != "LOL\n" {
    t.Fatalf("Expected the response to be \"LOL\\n\", got %s", rs)
  }
}
