package dockr

import (
  "encoding/json"
)

type VersionResponse struct {
  Version   string
  GitCommit string
  GoVersion string
}

func (c *Client) Version() (*VersionResponse, error) {
  res, err := c.callf("GET","/v1.4/version")
  if err != nil {
    return nil, err
  }
  if err = expectHTTPStatus(res.StatusCode, 200); err != nil {
    return nil, err
  }
  var rep VersionResponse
  dec := json.NewDecoder(res.Body)
  if err = dec.Decode(&rep) ; err != nil {
    return nil, err
  }
  return &rep, nil
}
