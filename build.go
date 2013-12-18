package dockr

import (
  "net/url"
  "net/http/httputil"
  "io"
  "bytes"
  "archive/tar"
  "bufio"
  "regexp"
)

type BuildRequest struct {
  Body            io.Reader
  Remote          string
  Tag             string
  Quiet           bool
  NoCache         bool
  RemoveTemporary bool
}

func (q *BuildRequest) Values() url.Values {
  values := url.Values{}
  values.Set("remote", q.Remote)
  values.Set("t", q.Tag)
  values.Set("q", boolString(q.Quiet))
  values.Set("nocache", boolString(q.NoCache))
  values.Set("rm", boolString(q.RemoveTemporary))
  return values
}

func (c *Client) Build(q *BuildRequest) (io.ReadCloser, error) {
  values := q.Values()
  res, client, err := c.callfquerybody2("POST","/v1.4/build", values, q.Body)
  if err != nil {
    if client != nil {
      client.Close()
    }
    return nil, err
  }
  if err = expectHTTPStatus(res.StatusCode, 200); err != nil {
    return nil, err
  }
  con, buf := client.Hijack()
  // response is chunked
  ch := httputil.NewChunkedReader(buf)
  return &hijackReadWriteCloser{con,ch}, nil
}

func SimpleDockerFile(content string) (io.Reader, error) {
  buf := new(bytes.Buffer)
  tw  := tar.NewWriter(buf)
  tw.WriteHeader(&tar.Header{Name:"Dockerfile", Size: int64(len(content))})
  tw.Write([]byte(content))
  tw.Close()
  return buf, nil
}

var (
  build_rx = regexp.MustCompile(`\A\r?Successfully built ([0-9a-f]{12})\z`)
)

func BuildStatusScanner(rd io.Reader) <-chan string {
  sc := bufio.NewScanner(rd)
  ch := make(chan string, 1)
  go func(){
    for sc.Scan() {
      txt := sc.Text()
      if m := build_rx.FindStringSubmatch(txt) ; m != nil {
        ch <- m[1]
        return
      }
    }
  }()
  return ch
}
