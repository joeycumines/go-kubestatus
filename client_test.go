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

import "testing"

func TestStatusOK(t *testing.T) {
	for _, testCase := range []struct {
		Status int
		OK     bool
	}{
		{0, false},
		{-1, false},
		{-200, false},
		{200, true},
		{299, true},
		{199, false},
		{300, false},
	} {
		if statusOK(testCase.Status) != testCase.OK {
			t.Error(testCase)
		}
	}
}

func TestClient_Get(t *testing.T) {
	for _, testCase := range []struct {
		URL    string
		Status bool
		Error  bool
	}{
		{"", false, true},
		{"https://httpbin.org", false, false},
		{"https://httpbin.org/anything", true, false},
		{"https://httpbin.org/status/418", false, true},
		{"https://httpbin.org/status/208", false, false},
	} {
		statuses, err := Client{
			Addresses: []string{testCase.URL},
		}.Get("")
		if (statuses[0] != nil) != testCase.Status {
			t.Error(statuses[0])
		}
		if (err != nil) != testCase.Error {
			t.Error(err)
		}
	}
}
