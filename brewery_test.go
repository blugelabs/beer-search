//  Copyright (c) 2020 The Bluge Authors.
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
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

func TestParseBrewery(t *testing.T) {
	expectBrewery := Brewery{
		Base: &Base{
			Type:    "brewery",
			Name:    "Yuengling & Son Brewing",
			Updated: DateTime(time.Date(2010, 7, 22, 20, 0, 20, 0, time.UTC)),
		},
		City:    "Pottsville",
		State:   "Pennsylvania",
		Country: "United States",
		Code:    "17901",
		Phone:   "570-622-0153",
		Website: "http://www.yuengling.com",
		Address: []string{"310 Mill Creek Avenue"},
		Geo: GeoPoint{
			Accuracy: "ROOFTOP",
			Lat:      40.7,
			Lon:      -76.1747,
		},
	}
	sampleBreweryData, err := ioutil.ReadFile("data/yuengling_son_brewing.json")
	if err != nil {
		t.Fatalf("error reading sample brewery json: %v", err)
	}
	var brewery Brewery
	err = json.Unmarshal(sampleBreweryData, &brewery)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(brewery, expectBrewery) {
		t.Errorf("expected beer: %#v, got: %#v", expectBrewery, brewery)
	}
}
