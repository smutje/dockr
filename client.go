package dockr

import (
  "fmt"
  "net"
  "net/http"
  "net/http/httputil"
  "strings"
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

func (c *Client) do(req *http.Request) (*http.Response, *httputil.ClientConn, error) {
  connection, err := c.connector.Connect()
  if err != nil {
    connection.Close()
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
/*
func (c *Client) simple(verb, url string) (*net.HTTPResponse, *httputil.ClientConn, err) {
}
*/
