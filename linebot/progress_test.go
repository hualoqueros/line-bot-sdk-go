// Copyright 2020 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package linebot

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

// TestGetProgressMessages tests GetProgressNarrowcastMessages func
func TestGetProgressMessages(t *testing.T) {
	type want struct {
		URLPath     string
		RequestBody []byte
		Response    *MessagesProgressResponse
		Error       error
	}
	var testCases = []struct {
		TestType     ProgressType
		RequestID    string
		ResponseCode int
		Response     []byte
		Want         want
	}{
		{
			RequestID:    "f70dd685-499a-1",
			TestType:     ProgressTypeNarrowcast,
			ResponseCode: 200,
			Response:     []byte(`{"phase":"waiting"}`),
			Want: want{
				URLPath: fmt.Sprintf(APIEndpointGetMessageProgress, ProgressTypeNarrowcast),
				Response: &MessagesProgressResponse{
					Phase:             "waiting",
					SuccessCount:      0,
					FailureCount:      0,
					TargetCount:       0,
					FailedDescription: "",
					ErrorCode:         0,
				},
			},
		},
		{
			RequestID:    "f70dd685-499a-2",
			TestType:     ProgressTypeNarrowcast,
			ResponseCode: 200,
			Response:     []byte(`{"phase":"succeeded","successCount":10,"failureCount":0,"targetCount":10}`),
			Want: want{
				URLPath: fmt.Sprintf(APIEndpointGetMessageProgress, ProgressTypeNarrowcast),
				Response: &MessagesProgressResponse{
					Phase:             "succeeded",
					SuccessCount:      10,
					FailureCount:      0,
					TargetCount:       10,
					FailedDescription: "",
					ErrorCode:         0,
				},
			},
		},
		{
			RequestID:    "f70dd685-499a-3",
			TestType:     ProgressTypeNarrowcast,
			ResponseCode: 200,
			Response:     []byte(`{"phase":"failed","failedDescription":"internal error","errorCode":1}`),
			Want: want{
				URLPath: fmt.Sprintf(APIEndpointGetMessageProgress, ProgressTypeNarrowcast),
				Response: &MessagesProgressResponse{
					Phase:             "failed",
					SuccessCount:      0,
					FailureCount:      0,
					TargetCount:       0,
					FailedDescription: "internal error",
					ErrorCode:         1,
				},
			},
		},
	}

	var currentTestIdx int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		tc := testCases[currentTestIdx]
		if r.Method != http.MethodGet {
			t.Errorf("Method %s; want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != tc.Want.URLPath {
			t.Errorf("URLPath %s; want %s", r.URL.Path, tc.Want.URLPath)
		}
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(tc.ResponseCode)
		w.Write(tc.Response)
	}))
	defer server.Close()

	dataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		t.Error("Unexpected data API call")
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not found"}`))
	}))
	defer dataServer.Close()

	client, err := mockClient(server, dataServer)
	if err != nil {
		t.Fatal(err)
	}
	var res interface{}
	for i, tc := range testCases {
		currentTestIdx = i
		t.Run(strconv.Itoa(i)+"/"+string(tc.TestType)+"."+tc.RequestID, func(t *testing.T) {
			switch tc.TestType {
			case ProgressTypeNarrowcast:
				res, err = client.GetProgressNarrowcastMessages(tc.RequestID).Do()
			}
			if tc.Want.Error != nil {
				if !reflect.DeepEqual(err, tc.Want.Error) {
					t.Errorf("Error %v; want %v", err, tc.Want.Error)
				}
			} else {
				if err != nil {
					t.Error(err)
				}
			}
			if tc.Want.Response != nil {
				if !reflect.DeepEqual(res, tc.Want.Response) {
					t.Errorf("Response %v; want %v", res, tc.Want.Response)
				}
			}
		})
	}
}
