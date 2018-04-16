/*
   Copyright 2018 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */

// Package status provides a simple, opinionated, kubernetes status http server.
package kubestatus

import (
	"fmt"
	"time"
	"errors"
	"net/url"
	"github.com/gin-gonic/gin"
)

const (
	DefaultPort      = 8080
	DefaultStartWait = time.Millisecond * 100
)

type (
	// HealthHandler should return an error if the service is not ready
	HealthHandler func() error

	// ReadinessHandler should return an error if the service is not ready
	ReadinessHandler func() error

	// Config provides configuration of the status http server.
	Config struct {
		// Port is the tcp port to serve the http server
		Port int

		// Hostname is the hostname fragment for the http server, which defaults to an empty string (all)
		Hostname string

		// StartWait is how long the kubestatus.Service.Start operation will block before checking health and readiness
		StartWait time.Duration

		// HealthHandler should return an error if the service is not ready
		HealthHandler HealthHandler

		// ReadinessHandler should return an error if the service is not ready
		ReadinessHandler ReadinessHandler

		// GinHandlers defines middleware to use
		GinHandlers []gin.HandlerFunc

		// Dependencies should be an array of addresses (including scheme) that are the root part of `/readiness`
		// endpoints, note that the `uuids` query parameter will be set, appending configured service's UUID on the
		// end of any existing `uuids` passed in with the original `/readiness` GET
		Dependencies []string
	}
)

// NewConfig creates a default config
func NewConfig() Config {
	return Config{
		Port:      DefaultPort,
		StartWait: DefaultStartWait,
		GinHandlers: []gin.HandlerFunc{
			gin.Logger(),
			gin.Recovery(),
		},
	}
}

// Validate returns an error if config is invalid
func (c Config) Validate() error {
	if c.Port < 0 {
		return fmt.Errorf("invalid port: %v", c.Port)
	}
	if c.StartWait < 0 {
		return fmt.Errorf("invalid start wait: %v", c.StartWait)
	}
	if c.HealthHandler == nil {
		return errors.New("nil HealthHandler")
	}
	if c.ReadinessHandler == nil {
		return errors.New("nil ReadinessHandler")
	}
	return nil
}

// URL returns the HTTP url this service will bind on (an empty host defaults to localhost), it has a http scheme
func (c Config) URL() string {
	URL := new(url.URL)
	URL.Scheme = "http"
	hostname := c.Hostname
	if hostname == "" {
		hostname = "localhost"
	}
	URL.Host = fmt.Sprintf("%s:%d", hostname, c.Port)
	return URL.String()
}
