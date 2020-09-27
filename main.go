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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/blugelabs/bluge/search"
	"github.com/blugelabs/bluge/search/aggregations"

	"github.com/blugelabs/bluge"
)

var batchSize = flag.Int("batchSize", 1000, "batch size for indexing")
var bindAddr = flag.String("addr", ":8094", "http listen address")
var jsonDir = flag.String("jsonDir", "data/", "json directory")
var beerIndexPath = flag.String("beerIndexPath", "beers.bluge", "beer index path")
var breweryIndexPath = flag.String("breweryIndexPath", "breweries.bluge", "brewery index path")
var staticPath = flag.String("static", "static/", "Path to the static content")
var doIndex = flag.Bool("index", true, "index or reindex the data")

var doTestSearch = flag.Bool("testSearch", false, "test search from another process")
var backupBeersTo = flag.String("backupBeersTo", "", "path to backup the beers index to")

func main() {
	flag.Parse()

	logger := log.New(os.Stderr, "beer-search", log.LstdFlags)

	fieldTypeBeer := bluge.NewKeywordField("_type", "beer").StoreValue().Aggregatable()
	beerCfg := bluge.DefaultConfig(*beerIndexPath).
		WithVirtualField(fieldTypeBeer)
	fieldTypeBrewery := bluge.NewKeywordField("_type", "brewery").StoreValue().Aggregatable()
	breweryCfg := bluge.DefaultConfig(*breweryIndexPath).
		WithVirtualField(fieldTypeBrewery)

	if *backupBeersTo != "" {
		indexReader, err := bluge.OpenReader(beerCfg)
		if err != nil {
			log.Fatalf("unable to open snapshot reader: %v", err)
		}
		err = indexReader.Backup(*backupBeersTo, nil)
		if err != nil {
			log.Fatalf("error backing up beers: %v", err)
		}
		return
	}

	if *doTestSearch {
		indexReader, err := bluge.OpenReader(beerCfg)
		if err != nil {
			log.Fatalf("unable to open snapshot reader: %v", err)
		}
		q := bluge.NewNumericRangeInclusiveQuery(0, bluge.MaxNumeric, false, true).SetField("abv")
		req := bluge.NewTopNSearch(0, q).WithStandardAggregations()
		styleAgg := aggregations.NewTermsAggregation(aggregations.FilterText(search.Field("style-facet"),
			func(bytes []byte) bool {
				return len(bytes) > 0
			}), 10)
		abvQuantile := aggregations.Quantiles(search.Field("abv"))
		styleAgg.AddAggregation("abvQuant", abvQuantile)
		req.AddAggregation(styleAggregation, styleAgg)
		dmi, err := indexReader.Search(context.Background(), req)
		if err != nil {
			log.Fatalf("error executing search: %v", err)
		}
		fmt.Printf("%d total hits\n", dmi.Aggregations().Count())
		styles := dmi.Aggregations().Aggregation(styleAggregation).(search.BucketCalculator)
		for _, styleBucket := range styles.Buckets() {
			abvQ := styleBucket.Aggregations()["abvQuant"].(*aggregations.QuantilesCalculator)
			p50, err := abvQ.Quantile(0.5)
			if err != nil {
				log.Fatal(err)
			}
			p99, err := abvQ.Quantile(0.99)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%35s - %4d - Median ABV: %4.1f 99%% ABV: %4.1f\n", styleBucket.Name(), styleBucket.Count(), p50, p99)
		}

		return
	}

	beerIndexWriter, err := bluge.OpenWriter(beerCfg)
	if err != nil {
		log.Fatalf("error opening beers index '%s': %v", *beerIndexPath, err)
	}

	breweryIndexWriter, err := bluge.OpenWriter(breweryCfg)
	if err != nil {
		log.Fatalf("error opening breweries index '%s': %v", *breweryIndexPath, err)
	}

	if *doIndex {
		go func() {
			err := indexData(beerIndexWriter, breweryIndexWriter)
			if err != nil {
				log.Fatalf("error indexing data: %v", err)
			}
		}()
	}
	// create a router to serve static files
	router := staticFileRouter()

	// add the API
	searchHandler := NewSearchHandler(beerIndexWriter, breweryIndexWriter, logger)
	router.Handle("/api/search", searchHandler).Methods("POST")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(*staticPath)))

	// start the HTTP server
	http.Handle("/", router)
	log.Printf("Listening on %v", *bindAddr)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}

func parseAndBuildDoc(dir, filename string) (Indexable, *bluge.Document, error) {
	obj, jsonBytes, err := parseJSONPath(dir, filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing JSON '%s': %w", filename, err)
	}
	doc, err := obj.Document(jsonBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error mapping object: %w", err)
	}
	return obj, doc, nil
}

func indexData(beerIndexWriter, breweryIndexWriter *bluge.Writer) error {
	log.Printf("Indexing...")
	startTime := time.Now()
	dirEntries, err := ioutil.ReadDir(*jsonDir)
	if err != nil {
		return err
	}

	var beerIndexedCount, breweryIndexedCount int
	var beers, breweries []*bluge.Document
	for _, dirEntry := range dirEntries {
		var obj Indexable
		var doc *bluge.Document
		obj, doc, err = parseAndBuildDoc(*jsonDir, dirEntry.Name())
		if err != nil {
			return err
		}
		switch obj.(type) {
		case *Beer:
			beers = append(beers, doc)
		case *Brewery:
			breweries = append(breweries, doc)
		}

		if len(beers) > *batchSize {
			err = indexBatch(beerIndexWriter, beers)
			if err != nil {
				return fmt.Errorf("error executing beer batch: %w", err)
			}
			beerIndexedCount += len(beers)
			beers = beers[:0]
		}
		if len(breweries) > *batchSize {
			err = indexBatch(breweryIndexWriter, breweries)
			if err != nil {
				return fmt.Errorf("error executing brewery batch: %w", err)
			}
			breweryIndexedCount += len(breweries)
			breweries = breweries[:0]
		}
	}
	if len(beers) > 0 {
		err = indexBatch(beerIndexWriter, beers)
		if err != nil {
			return fmt.Errorf("error executing beer batch: %w", err)
		}
		beerIndexedCount += len(beers)
	}
	if len(breweries) > 0 {
		err = indexBatch(breweryIndexWriter, breweries)
		if err != nil {
			return fmt.Errorf("error executing brewery batch: %w", err)
		}
		breweryIndexedCount += len(breweries)
	}

	indexTime := time.Since(startTime)
	timePerDoc := float64(indexTime) / float64(beerIndexedCount+breweryIndexedCount)
	log.Printf("Indexed %d documents, in %s (average %.2fms/doc)", beerIndexedCount+breweryIndexedCount,
		indexTime, timePerDoc/float64(time.Millisecond))
	return nil
}

func indexBatch(indexWriter *bluge.Writer, docs []*bluge.Document) error {
	batch := bluge.NewBatch()
	for _, doc := range docs {
		batch.Update(doc.ID(), doc)
	}
	return indexWriter.Batch(batch)
}
