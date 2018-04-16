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
