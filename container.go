package dockr

import (
  "io"
  "fmt"
  "net/url"
  "encoding/json"
)

/*
The following structs are @authored by docker.
*/

type CreateContainerRequest struct {
  Hostname        string
  User            string
  Memory          int64 // Memory limit (in bytes)
  MemorySwap      int64 // Total memory usage (memory + swap); set `-1' to disable swap
  CpuShares       int64 // CPU shares (relative weight vs. other containers)
  AttachStdin     bool
  AttachStdout    bool
  AttachStderr    bool
  //PortSpecs       []string
  ExposedPorts    map[string]struct{}
  Tty             bool // Attach standard streams to a tty, including stdin if it is not closed.
  OpenStdin       bool // Open stdin
  StdinOnce       bool // If true, close stdin after the 1 attached client disconnects.
  Env             []string
  Cmd             []string
  Dns             []string
  Image           string // Name of the image as it was passed by the operator (eg. could be symbolic)
  Volumes         map[string]struct{}
  VolumesFrom     string
  WorkingDir      string
  Entrypoint      []string
  NetworkDisabled bool
}

type CreateContainerResponse struct {
  Id              string
  Warnings        []string
}

type KeyValuePair struct {
  Key string
  Value string
}

type HostPort struct {
  Ip    string `json:"HostIp"`
  Port  string `json:"HostPort"`
}

type StartContainerRequest struct {
  Binds           []string
  ContainerIDFile string
  PortBindings    map[string][]HostPort
  LxcConf         []KeyValuePair
  Privileged      bool
}

type StopContainerRequest struct {
  Timeout         int
}

type Port struct {
  PrivatePort int64
  PublicPort  int64
  Type        string
  IP          string
}

func parsePort(key string, hp HostPort) Port {
  p := Port{IP: hp.Ip}
  fmt.Sscanf(key,"%d/%s", &p.PrivatePort, &p.Type)
  fmt.Sscanf(hp.Port,"%d",&p.PublicPort)
  return p
}

type fullContainer struct {
  XID              string    `json:"ID"`
  XImage           string    `json:"Image"`
  XHostConfig      StartContainerRequest `json:"HostConfig"`
}

type lightContainer struct {
  XID              string    `json:"ID"`
  XImage           string    `json:"Image"`
  XPorts           []Port    `json:"Ports"`
  full            *fullContainer `json:"-"`
}

func (l *fullContainer) Id() string {
  return l.XID
}
func (l *fullContainer) Image() string {
  return l.XImage
}
func (l *fullContainer) Ports() []Port {
  ports := make([]Port,0,len(l.XHostConfig.PortBindings))
  for k,binds := range(l.XHostConfig.PortBindings) {
    for _,p := range(binds){
      ports = append(ports, parsePort(k,p))
    }
  }
  return ports
}

func (l *lightContainer) Id() string {
  return l.XID
}
func (l *lightContainer) Image() string {
  return l.XImage
}
func (l *lightContainer) Ports() []Port {
  return l.XPorts
}


type Command struct {
  Executeable string
  Arguments   []string
}

type Container interface {
  Ports() []Port
  Image() string
}

type ExistingContainer interface {
  Id() string
  Container
}

func (q *StopContainerRequest) Values() url.Values {
  query := url.Values{}
  query.Set("t", fmt.Sprintf("%d",q.Timeout))
  return query
}

type AttachContainerRequest struct {
  // Which stream to attach:
  Stdin           bool
  Stdout          bool
  Stderr          bool
  // What to post on this streams:
  Logs            bool // get archived stuff?
  Stream          bool // stream new stuff?
}

type ListContainersRequest struct{
  All             bool
}

func boolString(b bool) string {
  if b {
    return "true"
  }else{
    return "false"
  }
}

func (q *AttachContainerRequest) Values() url.Values {
  query := url.Values{}
  query.Set("stdin" , boolString(q.Stdin) )
  query.Set("stdout", boolString(q.Stdout) )
  query.Set("stderr", boolString(q.Stderr))
  query.Set("logs"  , boolString(q.Logs))
  query.Set("stream", boolString(q.Stream))
  return query
}

func (q *ListContainersRequest) Values() url.Values {
  query := url.Values{}
  query.Set("all", boolString(q.All))
  return query
}

func (c *Client) CreateContainer(q *CreateContainerRequest) (*CreateContainerResponse, error){
  res, err := c.callfjson("POST","/v1.8/containers/create",q)
  if err != nil {
    return nil, err
  }
  err = expectHTTPStatus( res, 201 )
  if err != nil {
    return nil, err
  }
  var a CreateContainerResponse
  err = json.NewDecoder(res.Body).Decode(&a)
  if err != nil {
    return nil, err
  }
  return &a, nil
}

func (c *Client) DeleteContainer(id string) error {
  err := validateId(id)
  if err != nil {
    return err
  }
  res, err := c.callf("DELETE","/v1.8/containers/%s",id)
  if err != nil {
    return err
  }
  // 406 = you have to stop before delete
  res.Body.Close()
  return expectHTTPStatus(res, 204)
}

func (c *Client) StartContainer(id string, q *StartContainerRequest) error {
  err := validateId(id)
  if err != nil {
    return err
  }
  res, err := c.callfjson("POST","/v1.8/containers/%s/start",q, id)
  if err != nil {
    return err
  }
  res.Body.Close()
  return expectHTTPStatus(res, 204)
}

func (c *Client) StopContainer(id string, q *StopContainerRequest) error {
  err := validateId(id)
  if err != nil {
    return err
  }
  res, err := c.callfquery("POST","/v1.8/containers/%s/stop",q.Values(), id)
  if err != nil {
    return err
  }
  return expectHTTPStatus(res, 204)
}
func (c *Client) AttachContainer(id string, q *AttachContainerRequest) (io.ReadWriteCloser, error) {
  err := validateId(id)
  if err != nil {
    return nil, err
  }
  res, client, err := c.callfquery2("POST","/v1.4/containers/%s/attach",q.Values(), id)
  if err != nil {
    if client != nil {
      client.Close()
    }
    return nil, err
  }
  if err = expectHTTPStatus(res, 200); err != nil {
    return nil, err
  }
  con, buf := client.Hijack()
  return &hijackReadWriteCloser{con,buf}, nil
}

func (c *Client) GetContainer(id string) (ExistingContainer, error){
  err := validateId(id)
  if err != nil {
    return nil, err
  }
  res, err := c.callf("GET","/v1.8/containers/%s/json", id)
  if err != nil {
    return nil, err
  }
  defer res.Body.Close()
  if res.StatusCode == 404 {
    return nil, NOT_FOUND
  }
  if err = expectHTTPStatus(res, 200); err != nil {
    return nil, err
  }
  var a fullContainer
  err = json.NewDecoder(res.Body).Decode(&a)
  if err != nil {
    return nil, err
  }
  return &a, nil
}

func (c *Client) ListContainers(q *ListContainersRequest) ([]ExistingContainer, error){
  res, err := c.callfquery("GET","/v1.8/containers/json",q.Values())
  if err != nil {
    return nil, err
  }
  defer res.Body.Close()
  if err = expectHTTPStatus(res, 200); err != nil {
    return nil, err
  }
  var a []lightContainer;
  err = json.NewDecoder(res.Body).Decode(&a)
  if err != nil {
    return nil, err
  }
  r := make([]ExistingContainer,0,len(a))
  for _,cont := range(a) {
    lcont := cont
    r = append(r, &lcont)
  }
  return r, nil
}
