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

func TestParseBeer(t *testing.T) {
	expectBeer := Beer{
		Base: &Base{
			Type:        "beer",
			Name:        "Yuengling Lager",
			Description: "Yuengling Traditional Lager is an iconic American lager famous for its rich amber color and medium-bodied flavor. Brewed with roasted caramel malt for a subtle sweetness and a combination of cluster and cascade hops, this true original promises a well balanced taste with very distinct character. Its exceptional flavor and smooth finish is prevalently enjoyed by consumers with even the most discerning tastes. Our flagship brand, Yuengling Traditional Lager is an American favorite delivering consistent quality and refreshment that never disappoints. In fact, it's so widely known and unique in its class, in some areas you can ask for it simply by the name \"Lager.\"",
			Updated:     DateTime(time.Date(2010, 7, 22, 20, 0, 20, 0, time.UTC)),
		},
		BreweryID: "yuengling_son_brewing",
		ABV:       4.4,
		Style:     "American-Style Lager",
		Category:  "North American Lager",
	}
	sampleBeerData, err := ioutil.ReadFile("data/yuengling_son_brewing-yuengling_lager.json")
	if err != nil {
		t.Fatalf("error reading sample beer json: %v", err)
	}
	var beer Beer
	err = json.Unmarshal(sampleBeerData, &beer)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beer, expectBeer) {
		t.Errorf("expected beer: %#v, got: %#v", expectBeer, beer)
	}
}
