// This program converts a HAR (HTTP Archive) file into a Hoverfly simulation file.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
)

type HAR struct {
	Log struct {
		Entries []struct {
			Request struct {
				Method  string `json:"method"`
				URL     string `json:"url"`
				Headers []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"headers"`
				PostData struct {
					Text string `json:"text"`
				} `json:"postData"`
			} `json:"request"`
			Response struct {
				Status  int `json:"status"`
				Headers []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"headers"`
				Content struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"response"`
		} `json:"entries"`
	} `json:"log"`
}

type FieldMatcher struct {
	Matcher string `json:"matcher"`
	Value   string `json:"value"`
}

type Header map[string][]string

type Request struct {
	Method      []FieldMatcher            `json:"method"`
	Destination []FieldMatcher            `json:"destination"`
	Path        []FieldMatcher            `json:"path"`
	Body        []FieldMatcher            `json:"body,omitempty"`
	Headers     map[string][]FieldMatcher `json:"headers,omitempty"`
	Query       map[string][]FieldMatcher `json:"query,omitempty"`
}

type Response struct {
	Status  int    `json:"status"`
	Body    string `json:"body,omitempty"`
	Headers Header `json:"headers,omitempty"`
}

type Pair struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Labels   []string `json:"labels"`
}

type GlobalActions struct {
	Delays []string `json:"delays"`
}

type Simulation struct {
	Data struct {
		Pairs         []Pair        `json:"pairs"`
		GlobalActions GlobalActions `json:"globalActions"`
	} `json:"data"`
	Meta struct {
		SchemaVersion string `json:"schemaVersion"`
	} `json:"meta"`
}

func isTextContentType(contentType string) bool {
	lower := strings.ToLower(contentType)
	return strings.Contains(lower, "json") || strings.Contains(lower, "text") || strings.Contains(lower, "xml")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func main() {
	outputFile := flag.String("output", "", "Output file path (optional, defaults to stdout)")
	maxResponseSize := flag.Int("max-response-bytes", -1, "Max response body size in bytes (optional, default: unlimited)")
	skipNonText := flag.Bool("skip-non-text", false, "If set, non-text content types will get a generic response body")
	hostFilter := flag.String("host", "", "Only include entries with this destination host (optional)")
	summarise := flag.Bool("summarise", false, "If set, summarises request/response pairs instead of generating a simulation file")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: har-to-hoverfly [--output file.json] [--max-response-bytes N] [--skip-non-text] [--host hostname] [--summarise] <input.har>")
		os.Exit(1)
	}

	inputFile := flag.Arg(0)
	harData, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	var har HAR
	err = json.Unmarshal(harData, &har)
	if err != nil {
		log.Fatal(err)
	}

	if *summarise {
		summaryMap := make(map[string][]string)
		for _, entry := range har.Log.Entries {
			r := entry.Request
			u, err := url.Parse(r.URL)
			if err != nil {
				continue
			}
			if *hostFilter != "" && u.Host != *hostFilter {
				continue
			}
			method := truncate(r.Method, 10)
			path := truncate(u.Path, 50)
			query := truncate(u.RawQuery, 50)
			summary := fmt.Sprintf("%-10s %-50s %-50s", method, path, query)
			summaryMap[u.Host] = append(summaryMap[u.Host], summary)
		}
		for host, summaries := range summaryMap {
			fmt.Printf("\nHOST: %s\n", host)
			fmt.Printf("  %-10s %-50s %-50s\n", "METHOD", "PATH", "QUERY")
			sort.Strings(summaries)
			for _, s := range summaries {
				fmt.Println("  ", s)
			}
		}

		return
	}
	var sim Simulation
	sim.Meta.SchemaVersion = "v5.3"
	sim.Data.GlobalActions = GlobalActions{Delays: []string{}}

	for _, entry := range har.Log.Entries {
		r := entry.Request
		rURL := r.URL
		u, err := url.Parse(rURL)
		if err != nil {
			continue
		}

		if *hostFilter != "" && u.Host != *hostFilter {
			continue
		}

		// Request headers
		reqHeaders := make(map[string][]FieldMatcher)
		for _, h := range r.Headers {
			reqHeaders[h.Name] = append(reqHeaders[h.Name], FieldMatcher{Matcher: "exact", Value: h.Value})
		}

		// Request query parameters
		reqQuery := make(map[string][]FieldMatcher)
		for key, values := range u.Query() {
			for _, v := range values {
				reqQuery[key] = append(reqQuery[key], FieldMatcher{Matcher: "exact", Value: v})
			}
		}

		// Response headers
		contentType := ""
		headersMap := make(Header)
		for _, h := range entry.Response.Headers {
			lower := strings.ToLower(h.Name)
			if lower == "content-type" {
				contentType = h.Value
			}
			headersMap[h.Name] = append(headersMap[h.Name], h.Value)
		}

		responseBody := entry.Response.Content.Text
		if *maxResponseSize >= 0 && len(responseBody) > *maxResponseSize {
			responseBody = ""
		}
		if *skipNonText && !isTextContentType(contentType) {
			responseBody = "NON_TEXT_RESPONSE_SKIPPED"
		}

		pair := Pair{
			Request: Request{
				Method:      []FieldMatcher{{Matcher: "exact", Value: r.Method}},
				Destination: []FieldMatcher{{Matcher: "exact", Value: u.Host}},
				Path:        []FieldMatcher{{Matcher: "exact", Value: u.Path}},
				Headers:     reqHeaders,
				Query:       reqQuery,
			},
			Response: Response{
				Status:  entry.Response.Status,
				Body:    responseBody,
				Headers: headersMap,
			},
			Labels: []string{},
		}

		if r.PostData.Text != "" {
			pair.Request.Body = []FieldMatcher{{Matcher: "exact", Value: r.PostData.Text}}
		}

		sim.Data.Pairs = append(sim.Data.Pairs, pair)
	}

	output, err := json.MarshalIndent(sim, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if *outputFile != "" {
		err := os.WriteFile(*outputFile, output, 0644)
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	} else {
		fmt.Println(string(output))
	}
}
