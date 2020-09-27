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
	"time"

	"github.com/blugelabs/bluge"
)

type Base struct {
	ID          string   `json:"-"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Updated     DateTime `json:"updated,omitempty"`
}

func (b *Base) Identifier() bluge.Identifier {
	return bluge.Identifier(b.ID)
}

func (b *Base) Document(jsonBytes []byte) *bluge.Document {
	doc := bluge.NewDocument(b.ID).
		AddField(bluge.NewStoredOnlyField("_source", jsonBytes)).
		AddField(bluge.NewKeywordField("type", b.Type)).
		AddField(bluge.NewTextField("name", b.Name)).
		AddField(bluge.NewTextField("desc", b.Description).SearchTermPositions()).
		AddField(bluge.NewDateTimeField("updated", time.Time(b.Updated)))
	return doc
}

const rfc3339NoTimezoneNoT = "2006-01-02 15:04:05"

type DateTime time.Time

func (d DateTime) MarshalJSON() ([]byte, error) {
	timeStr := time.Time(d).Format(rfc3339NoTimezoneNoT)
	return json.Marshal(timeStr)
}

func (d *DateTime) UnmarshalJSON(data []byte) error {
	// parse date/time string out of JSON
	var dateTimeStr string
	err := json.Unmarshal(data, &dateTimeStr)
	if err != nil {
		return fmt.Errorf("error parsing date/time to string: %w", err)
	}
	// parse date/time string as time.Time
	t, err := time.Parse(rfc3339NoTimezoneNoT, dateTimeStr)
	if err != nil {
		return fmt.Errorf("error parsing date/time string as time.Time: %w", err)
	}
	*d = DateTime(t)
	return nil
}
