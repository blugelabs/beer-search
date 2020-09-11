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

// the URL of your hosted search index
var searchURL = 'http://localhost:8094/api/search'

var filters = [];

$(document).ready(function () {

    // search result partials
    var searchResultTmpl = Handlebars.compile($('#searchResultTmpl').html());
    Handlebars.registerPartial('searchResultTmpl', searchResultTmpl);
    var beerTmpl = Handlebars.compile($('#beer').html());
    Handlebars.registerPartial('beer', beerTmpl);
    var breweryTmpl = Handlebars.compile($('#brewery').html());
    Handlebars.registerPartial('brewery', breweryTmpl);
    var explanationTmpl = Handlebars.compile($('#searchResultExplanationTmpl').html());
    Handlebars.registerPartial('searchResultExplanationTmpl', explanationTmpl);

    // aggregation partials
    var aggregationTmpl = Handlebars.compile($('#aggregationTmpl').html());
    Handlebars.registerPartial('aggregationTmpl', aggregationTmpl);

    // main templates
    var searchResultsTmpl = Handlebars.compile($('#searchResultsTmpl').html());
    var aggregationsTmpl = Handlebars.compile($('#aggregationsTmpl').html());


    Handlebars.registerHelper('documentType', function(context, options) {
        return this.document.type;
    });

    Handlebars.registerHelper('roundScore', function(number) {
        return roundScore(number);
    });

    $("#searchForm").submit(function() {
        newq = $("#query").val();
        if (newq !== userQuery) {
            // reset to first page
            $("#page").val(1);
            // reset filters
            $('input:checkbox').removeAttr('checked');
        }
    });


    parseFilters();
    var url = new URL(window.location.href);
    var userQuery = url.searchParams.get("q");
    console.log(userQuery)

    var page = getURIParameter("p", false);
    console.log("see p param as", page);
    if (!page) {
        page = 1;
    }
    $("#page").val(page);

    if (userQuery) {
        $("#query").val(userQuery);
        data = {
            "query": userQuery,
            "filters": filters,
            "page": parseInt(page),
        }
        $.ajax({
            type: "POST",
            url: searchURL,
            processData: false,
            contentType: 'application/json',
            data: JSON.stringify(data),
            success: function(r) {
                console.log(r);
                console.log("filters still", filters);
                r.filters = filters;
                $('#searchResultsArea').html(searchResultsTmpl(r));
                $('#aggregationsArea').html(aggregationsTmpl(r));
            },
            error: function(jqxhr, text, error) {
                console.log(error);
            }
        });
    }

})

function resubmit() {
    $("#page").val(1);
    $("#searchForm").submit();
}

var facets = ["type", "style-facet", "updated", "abv"];

function parseFilters() {
    for (var fnamei in facets) {
        var fname = facets[fnamei];
        console.log("looking for ", "f_"+fname);
        let fvals = getURIParameter("f_"+fname, true);
        console.log("see fvals", fvals);
        for (var fvi in fvals) {
            var fv = fvals[fvi];
            filters.push({"name": fname, "value": fv});
        }
    }
    console.log("see filters", filters);
}

function getURIParameter(param, asArray) {
    return document.location.search.substring(1).split('&').reduce(function(p,c) {
        var parts = c.split('=', 2).map(function(param) { return decodeURIComponent(param).replace(/\+/g, " "); });
        if(parts.length === 0 || parts[0] != param) return (p instanceof Array) && !asArray ? null : p;
        return asArray ? p.concat(parts.concat(true)[1]) : parts.concat(true)[1];
    }, []);
}

// set the page, resubmit the form
function jumpToPage(page, e) {
    if (e) {
        e.preventDefault();
    }
    $("#page").val(""+page);
    $("#searchForm").submit();
    return false;
}

function toggleScore(id, e) {
    console.log("toggle score");
    if (e) {
        e.preventDefault();
    }
    console.log("toggling", id);
    $("#score-"+id).toggle();
    return false;
}

function roundScore(score) {
    return Math.round(score*1000)/1000;
}