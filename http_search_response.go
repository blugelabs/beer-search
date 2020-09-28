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
	"math"

	"github.com/blugelabs/bluge/search"
)

type DocumentMatch struct {
	Document interface{}         `json:"document"`
	Score    float64             `json:"score"`
	Expl     *search.Explanation `json:"explanation"`
	ID       string              `json:"id"`
}

type AggregationValue struct {
	DisplayName string `json:"display_name"`
	FilterName  string `json:"filter_name"`
	Count       uint64 `json:"count"`
	Filtered    bool   `json:"filtered"`
}

type Aggregation struct {
	DisplayName string              `json:"display_name"`
	FilterName  string              `json:"filter_name"`
	Values      []*AggregationValue `json:"values"`
}

type SearchResponse struct {
	Query        string                  `json:"query"`
	Total        uint64                  `json:"total"`
	TopScore     float64                 `json:"top_score"`
	Hits         []*DocumentMatch        `json:"hits"`
	Duration     string                  `json:"duration"`
	Aggregations map[string]*Aggregation `json:"aggregations"`
	Message      string                  `json:"message"`
	PreviousPage int                     `json:"previousPage,omitempty"`
	NextPage     int                     `json:"nextPage,omitempty"`
}

func NewSearchResponse(query string) *SearchResponse {
	return &SearchResponse{
		Query: query,
	}
}

func (s *SearchResponse) buildAggregation(aggs *search.Bucket, name string, filters []*Filter) {
	agg := &Aggregation{
		DisplayName: displayName(name),
		FilterName:  name,
	}

	for _, bucket := range aggs.Buckets(name) {
		aggVal := &AggregationValue{
			DisplayName: displayName(bucket.Name()),
			FilterName:  bucket.Name(),
			Count:       bucket.Count(),
		}
		for _, f := range filters {
			if f.Name == name && f.Value == bucket.Name() {
				aggVal.Filtered = true
			}
		}
		agg.Values = append(agg.Values, aggVal)
	}

	s.Aggregations[name] = agg
}

func (s *SearchResponse) AddAggregations(aggs *search.Bucket, filters []*Filter) {
	s.Total = aggs.Count()
	s.TopScore = aggs.Metric("max_score")
	s.Duration = aggs.Duration().String()

	s.Aggregations = make(map[string]*Aggregation)
	s.buildAggregation(aggs, typeAggregation, filters)
	s.buildAggregation(aggs, styleAggregation, filters)
	s.buildAggregation(aggs, updatedAggregation, filters)
	s.buildAggregation(aggs, abvAggregation, filters)
}

func (s *SearchResponse) AddPaging(aggs *search.Bucket, page int) {
	numPages := int(math.Ceil(float64(aggs.Count()) / float64(resultsPerPage)))
	if numPages > page {
		s.NextPage = page + 1
	}
	if page != 1 {
		s.PreviousPage = page - 1
	}

	if page != 1 {
		s.Message = fmt.Sprintf("Page %d of ", page)
	}
	s.Message += fmt.Sprintf("%d results (%s)", aggs.Count(),
		aggs.Duration().Round(roundDurationTo))
}

func displayName(in string) string {
	switch in {
	case typeAggregation:
		return "Type"
	case "beer":
		return "Beer"
	case "brewery":
		return "Brewery"
	case updatedAggregation:
		return "Updated"
	case "old":
		return "Before 2012"
	case "new":
		return "Since 2012"
	case styleAggregation:
		return "Style"
	case abvAggregation:
		return "ABV"
	case "low":
		return "Low (< 3%)"
	case "med":
		return "Medium (3% - 5%)"
	case "high":
		return "High (> 5%)"
	}
	return in
}
