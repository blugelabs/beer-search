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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/blugelabs/bluge"
)

const typeBrewery = "brewery"
const typeBeer = "beer"

type Indexable interface {
	Identifier() bluge.Identifier
	Document([]byte) (*bluge.Document, error)
}

func parseJSONPath(dir, filename string) (Indexable, []byte, error) {
	docID := filename[:(len(filename) - len(filepath.Ext(filename)))]
	jsonBytes, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file '%s': %v", filename, err)
	}
	if strings.Contains(filename, "-") {
		return unmarshalByType("beer", docID, jsonBytes)
	}
	return unmarshalByType("brewery", docID, jsonBytes)
}

func unmarshalByType(_type, _id string, _source []byte) (rv Indexable, src []byte, err error) {
	switch _type {
	case typeBeer:
		rv = NewBeer(_id)
	case typeBrewery:
		rv = NewBrewery(_id)
	default:
		return nil, nil, fmt.Errorf("unsupported type: %s", _type)
	}
	err = json.Unmarshal(_source, rv)
	if err != nil {
		return nil, nil, err
	}
	return rv, _source, nil
}
