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
	"fmt"
	"log"
	"math"
	"time"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"

	"github.com/blugelabs/bluge"
	querystr "github.com/blugelabs/query_string"
)

func mustTimeParse(format, val string) time.Time {
	tt, err := time.Parse(format, val)
	if err != nil {
		panic(err)
	}
	return tt
}

const midDate = "2011-01-06T00:00:00Z"

var updatedRanges = map[string]*struct {
	Start time.Time
	End   time.Time
}{
	"old": {
		Start: time.Time{},
		End:   mustTimeParse(time.RFC3339, midDate),
	},
	"new": {
		Start: mustTimeParse(time.RFC3339, midDate),
		End:   time.Time{},
	},
}

var abvRanges = map[string]struct {
	Low  float64
	High float64
}{
	"low": {
		Low:  0,
		High: 3,
	},
	"med": {
		Low:  3,
		High: 5,
	},
	"high": {
		Low:  5,
		High: math.Inf(1),
	},
}

type Filter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SearchRequest struct {
	Query   string    `json:"query"`
	Filters []*Filter `json:"filters"`
	Page    int       `json:"page"`
}

func (r *SearchRequest) buildFilterClauses() (rv []bluge.Query) {
	for _, filter := range r.Filters {
		switch filter.Name {
		case typeAggregation, styleAggregation:
			rv = append(rv, bluge.NewTermQuery(filter.Value).SetField(filter.Name))
		case abvAggregation:
			if abvRange, ok := abvRanges[filter.Value]; ok {
				rv = append(rv, bluge.NewNumericRangeQuery(abvRange.Low, abvRange.High).SetField(filter.Name))
			}
		case updatedAggregation:
			if updatedRange, ok := updatedRanges[filter.Value]; ok {
				rv = append(rv, bluge.NewDateRangeQuery(updatedRange.Start, updatedRange.End).SetField(filter.Name))
			}
		}

		log.Printf("see filter name: %s value: %s", filter.Name, filter.Value)
	}

	return rv
}

func (r *SearchRequest) SizeOffset() (size, offset int) {
	return resultsPerPage, (r.Page - 1) * resultsPerPage
}

func (r *SearchRequest) BlugeRequest() (bluge.SearchRequest, error) {
	userQuery, err := querystr.ParseQueryString(r.Query, querystr.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("errror parsing query string '%s': %v", r.Query, err)
	}

	if r.Page < 1 {
		r.Page = 1
	}

	size, offset := r.SizeOffset()

	filters := r.buildFilterClauses()

	q := bluge.NewBooleanQuery().
		AddMust(userQuery).
		AddMust(filters...)

	blugeRequest := bluge.NewTopNSearch(size, q).
		WithStandardAggregations().
		SetFrom(offset).
		ExplainScores()

	blugeRequest.AddAggregation(typeAggregation, aggregations.NewTermsAggregation(search.Field("_type"), 2))
	styleAgg := aggregations.NewTermsAggregation(aggregations.FilterText(search.Field("style-facet"),
		func(bytes []byte) bool {
			return len(bytes) > 0
		}), 5)
	blugeRequest.AddAggregation(styleAggregation, styleAgg)

	updatedAgg := aggregations.DateRanges(search.Field("updated"))
	for k, v := range updatedRanges {
		log.Printf("start %v end %v", v.Start, v.End)
		updatedAgg.AddRange(aggregations.NewNamedDateRange(k, v.Start, v.End))
	}
	blugeRequest.AddAggregation(updatedAggregation, updatedAgg)

	abvAgg := aggregations.Ranges(search.Field("abv"))
	for k, v := range abvRanges {
		abvAgg.AddRange(aggregations.NamedRange(k, v.Low, v.High))
	}
	blugeRequest.AddAggregation(abvAggregation, abvAgg)

	return blugeRequest, nil
}
