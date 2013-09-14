package dockr

import (
  "fmt"
  "net"
  "net/http"
  "net/url" 
  "net/http/httputil"
  "strings"
  "bytes"
  "encoding/json"
)

type Connector interface {
  Connect() (net.Conn, error);
  fmt.Stringer
}

type DefaultConnector struct {
  network, address string
}

func (c *DefaultConnector) Connect() (net.Conn, error) {
  return net.Dial(c.network, c.address)
}

func (c *DefaultConnector) String() string {
  return fmt.Sprintf("%s://%s", c.network, c.address)
}

func NewDefaultConnector( endpoint string ) (*DefaultConnector, error) {
  subs := strings.SplitN( endpoint, "://", 2 )
  if len(subs) != 2 {
    return nil, fmt.Errorf("Enpoint must be given as <protocol>://<address> (e.g. unix:///var/run/docker.sock or https://localhost:4243)")
  }
  return &DefaultConnector{ subs[0], subs[1] }, nil
}

type Client struct {
  connector Connector
}

func NewClient( endpoint string ) (*Client, error) {
  con, err := NewDefaultConnector(endpoint)
  if err != nil {
    return nil, err
  }
  return &Client{con}, nil
}

func NewClientFromConnector( con Connector ) *Client {
  return &Client{con}
}

func (c *Client) do2(req *http.Request) (*http.Response, *httputil.ClientConn, error) {
  connection, err := c.connector.Connect()
  if err != nil {
    if connection != nil {
      connection.Close()
    }
    return nil, nil, err
  }
  con := httputil.NewClientConn( connection, nil )
  res, err := con.Do( req )
  // This error is set if the response didn't indicate 
  // explicit connection persistance. Docker does mean 
  // things with implicitly persistant connections, so 
  // ignore it.
  if err == httputil.ErrPersistEOF {
    err = nil
  }
  return res, con, err
}

func (c *Client) do(req *http.Request) (res *http.Response, err error){
  res, con, err := c.do2(req)
  if con != nil {
    con.Close()
  }
  return
}

func (c *Client) callf2(verb, format string, args ...string) (*http.Response, *httputil.ClientConn, error){
  eargs := make([]interface{}, len(args))
  for i,arg := range args {
    eargs[i] = url.QueryEscape(arg)
  }
  req, err := http.NewRequest(verb, fmt.Sprintf(format, eargs...), nil)
  if err != nil {
    return nil, nil, err
  }
  return c.do2(req)
}

func (c *Client) callf(verb, format string, args ...string) (res *http.Response, err error){
  res, con, err := c.callf2(verb, format, args... )
  if con != nil {
    con.Close()
  }
  return
}

func (c *Client) callfjson2(verb, format string, body interface{}, args ...string) (*http.Response, *httputil.ClientConn, error){
  eargs := make([]interface{}, len(args))
  for i,arg := range args {
    eargs[i] = url.QueryEscape(arg)
  }
  buf, err := json.Marshal(body)
  if err != nil {
    return nil, nil, err
  }
  req, err := http.NewRequest(verb, fmt.Sprintf(format, eargs...), bytes.NewBuffer(buf))
  req.Header.Set("Content-Type","application/json")
  if err != nil {
    return nil, nil, err
  }
  return c.do2(req)
}

func (c *Client) callfjson(verb, format string, body interface{}, args ...string) (res *http.Response,err  error){
  res, con, err := c.callfjson2(verb, format, body, args...)
  if con != nil {
    con.Close()
  }
  return
}

func (c *Client) callfquery2(verb, format string, query url.Values, args ...string) (*http.Response, *httputil.ClientConn, error){
  eargs := make([]interface{}, len(args))
  for i,arg := range args {
    eargs[i] = url.QueryEscape(arg)
  }
  url := strings.Join([]string{
    fmt.Sprintf(format, eargs...),
    query.Encode(),
  },"?")
  req, err := http.NewRequest(verb,url,nil)
  if err != nil {
    return nil, nil, err
  }
  return c.do2(req)
}

func (c *Client) callfquery(verb, format string, query url.Values, args ...string) (res *http.Response, err error){
  res, con, err := c.callfquery2(verb, format, query, args...)
  if con != nil {
    con.Close()
  }
  return
}

