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
	"github.com/blugelabs/bluge"
)

type GeoPoint struct {
	Accuracy string  `json:"accuracy"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
}

type Brewery struct {
	*Base
	City    string   `json:"city"`
	State   string   `json:"state"`
	Country string   `json:"country"`
	Code    string   `json:"code"`
	Phone   string   `json:"phone"`
	Website string   `json:"website"`
	Address []string `json:"address"`
	Geo     GeoPoint ` json:"geo"`
}

func NewBrewery(id string) *Brewery {
	return &Brewery{
		Base: &Base{
			ID: id,
		},
	}
}

func (b *Brewery) Document(jsonBytes []byte) (*bluge.Document, error) {
	doc := b.Base.Document(jsonBytes)

	doc.AddField(bluge.NewTextField("city", b.City))
	doc.AddField(bluge.NewTextField("state", b.State))
	doc.AddField(bluge.NewTextField("country", b.Country))
	doc.AddField(bluge.NewKeywordField("code", b.Code))
	doc.AddField(bluge.NewKeywordField("phone", b.Phone))
	doc.AddField(bluge.NewTextField("website", b.Website))
	for _, addr := range b.Address {
		doc.AddField(bluge.NewTextField("address", addr).SearchTermPositions())
	}
	doc.AddField(bluge.NewGeoPointField("location", b.Geo.Lon, b.Geo.Lat))

	doc.AddField(bluge.NewCompositeFieldIncluding("_all", []string{"name", "desc", "city", "state", "country", "address"}))

	return doc, nil
}
