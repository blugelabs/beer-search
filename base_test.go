//  Copyright (c) 2020 Bluge Labs, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name   string
		in     []byte
		expect time.Time
	}{
		{
			name:   "simple",
			in:     []byte(`"2010-07-22 20:00:20"`),
			expect: time.Date(2010, 7, 22, 20, 0, 20, 0, time.UTC),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var d DateTime
			err := json.Unmarshal(test.in, &d)
			if err != nil {
				t.Fatalf("error unmarshaling json: %v", err)
			}
			dt := time.Time(d)
			if !dt.Equal(test.expect) {
				t.Errorf("expected date time: %v got: %v", test.expect, dt)
			}
		})
	}
}
