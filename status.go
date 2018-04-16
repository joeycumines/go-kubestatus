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

package kubestatus

import (
	"time"
	guuid "github.com/google/uuid"
	"net/http"
)

// Status is the response object returned by all endpoints
type Status struct {
	// Code is the HTTP status code
	Code int `json:"code"`

	// Message will be either 'OK', or the error message
	Message string `json:"message"`

	// Success will be bool set to false for anything but 200
	Success bool `json:"success"`

	// Started is a nanoseconds epoch indicating when the service was started
	Started int64 `json:"started"`

	// Uptime is a human readable string representation of the current timestamp - started
	Uptime string `json:"uptime"`

	// UUID is a per-process uuid value in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	UUID string `json:"uuid"`
}

// NewStatus creates a new Status
func NewStatus(uuid [16]byte, started time.Time, err error) Status {
	startedTS := started.UnixNano()
	result := Status{
		Code:    http.StatusOK,
		Message: "OK",
		Success: true,
		Started: startedTS,
		Uptime:  time.Duration(time.Now().UnixNano() - startedTS).String(),
		UUID:    guuid.UUID(uuid).String(),
	}
	if err != nil {
		result.Code = http.StatusServiceUnavailable
		result.Message = err.Error()
		result.Success = false
	}
	return result
}
