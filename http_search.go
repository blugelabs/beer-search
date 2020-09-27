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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
)

const resultsPerPage = 10
const roundDurationTo = 500 * time.Microsecond
const styleAggregation = "style-facet"
const abvAggregation = "abv"
const typeAggregation = "type"
const updatedAggregation = "updated"

// SearchHandler can handle search requests sent over HTTP
type SearchHandler struct {
	beerIndexWriter    *bluge.Writer
	breweryIndexWriter *bluge.Writer
	logger             *log.Logger
}

func NewSearchHandler(beerIndexWriter, breweryIndexWriter *bluge.Writer, logger *log.Logger) *SearchHandler {
	return &SearchHandler{
		beerIndexWriter:    beerIndexWriter,
		breweryIndexWriter: breweryIndexWriter,
		logger:             logger,
	}
}

func (h *SearchHandler) Readers() (beerReader, breweryReader *bluge.Reader, err error) {
	beerReader, err = h.beerIndexWriter.Reader()
	if err != nil {
		return nil, nil, err
	}
	breweryReader, err = h.breweryIndexWriter.Reader()
	if err != nil {
		return nil, nil, err
	}
	return beerReader, breweryReader, nil
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		showError(w, req, fmt.Sprintf("error reading request body: %v", err), 400, h.logger)
		return
	}

	var searchRequest SearchRequest
	err = json.Unmarshal(requestBody, &searchRequest)
	if err != nil {
		showError(w, req, fmt.Sprintf("error parsing request: %v", err), 400, h.logger)
		return
	}

	blugeRequest, err := searchRequest.BlugeRequest()
	if err != nil {
		showError(w, req, err.Error(), 400, h.logger)
		return
	}

	beerReader, breweryReader, err := h.Readers()
	if err != nil {
		showError(w, req, err.Error(), 400, h.logger)
		return
	}

	blugeResponse, err := bluge.MultiSearch(context.Background(), blugeRequest, beerReader, breweryReader)
	if err != nil {
		showError(w, req, fmt.Sprintf("error executing query: %v", err), 500, h.logger)
		return
	}

	searchResponse := NewSearchResponse(searchRequest.Query)

	next, err := blugeResponse.Next()
	for err == nil && next != nil {
		var docID string
		var doc Indexable
		docID, doc, err = matchToIndexable(next)
		if err != nil {
			showError(w, req, fmt.Sprintf("error restoring document from match: %v", err), 500, h.logger)
			return
		}

		searchResponse.Hits = append(searchResponse.Hits, &DocumentMatch{
			ID:       docID,
			Document: doc,
			Score:    next.Score,
			Expl:     next.Explanation,
		})

		next, err = blugeResponse.Next()
	}
	if err != nil {
		showError(w, req, fmt.Sprintf("error executing query: %v", err), 500, h.logger)
		return
	}

	searchResponse.AddAggregations(blugeResponse.Aggregations(), searchRequest.Filters)
	searchResponse.AddPaging(blugeResponse.Aggregations(), searchRequest.Page)

	mustEncode(w, searchResponse)
}

func matchToIndexable(d *search.DocumentMatch) (string, Indexable, error) {
	var _id string
	var _type string
	var _source []byte
	err := d.VisitStoredFields(func(field string, value []byte) bool {
		switch field {
		case "_id":
			_id = string(value)
		case "_type":
			_type = string(value)
		case "_source":
			_source = value
		}
		return true
	})
	if err != nil {
		return "", nil, fmt.Errorf("error visiting stored fields: %v", err)
	}

	doc, _, err := unmarshalByType(_type, _id, _source)
	if err != nil {
		return "", nil, fmt.Errorf("error unmarshaling source: %v", err)
	}
	return _id, doc, nil
}
