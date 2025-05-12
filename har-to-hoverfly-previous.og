package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type HAR struct {
	Log struct {
		Entries []Entry `json:"entries"`
	} `json:"log"`
}

type Entry struct {
	StartedDateTime time.Time `json:"startedDateTime"`
	Time            float64   `json:"time"`
	Request         struct {
		Method string `json:"method"`
		URL    string `json:"url"`
	} `json:"request"`
	Response struct {
		Status  int `json:"status"`
		Content struct {
			MimeType string `json:"mimeType"`
			Text     string `json:"text"`
		} `json:"content"`
	} `json:"response"`
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

func main() {
	inputFile := flag.String("input", "", "Path to HAR file")
	outputFile := flag.String("output", "", "Path to output simulation JSON file (optional)")
	sizeLimit := flag.Int("max-body-bytes", 0, "Optional maximum body size (in bytes). Larger responses will be replaced with an empty body.")
	ignoreNonText := flag.Bool("ignore-non-text", false, "If set, non-textual content types will be excluded entirely from the simulation")
	allowedTypes := flag.String("allowed-content-types", "json,xml,text/html,text/javascript", "Comma-separated list of MIME substrings considered text-based")
	restrictHost := flag.String("host", "", "Restrict to entries for this destination host only")
	summarise := flag.Bool("summarise", false, "Summarise request/response pairs grouped by host")
	flag.Parse()

	allowedContentTypes := strings.Split(*allowedTypes, ",")

	if *inputFile == "" {
		log.Fatal("You must provide a HAR file with --input")
	}

	data, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var har HAR
	err = json.Unmarshal(data, &har)
	if err != nil {
		log.Fatalf("Failed to parse HAR: %v", err)
	}

	sim := Simulation{}
	sim.Meta.SchemaVersion = "v5.3"
	sim.Data.GlobalActions = GlobalActions{Delays: []string{}}

	table := make(map[string]map[string]map[string]bool)

	for _, entry := range har.Log.Entries {
		req := entry.Request
		res := entry.Response
		reqURL := parseURL(req.URL)

		if *restrictHost != "" {
			if !strings.Contains(req.URL, *restrictHost) {
				continue
			}
		}

		isText := isTextContent(res.Content.MimeType, allowedContentTypes)
		if *ignoreNonText && !isText {
			continue
		}

		if *summarise {
			host := reqURL.Host
			if _, ok := table[host]; !ok {
				table[host] = make(map[string]map[string]bool)
			}
			if _, ok := table[host][reqURL.Path]; !ok {
				table[host][reqURL.Path] = make(map[string]bool)
			}
			table[host][reqURL.Path][req.Method] = true
			continue
		}

		pair := convertEntryToPair(entry, *sizeLimit, allowedContentTypes)
		sim.Data.Pairs = append(sim.Data.Pairs, pair)
	}

	if *summarise {
		fmt.Printf("%-30s %-10s %-50s %-50s\n", "HOST", "METHOD", "PATH", "QUERY")
		for host, paths := range table {
			for path, methods := range paths {
				for method := range methods {
					fmt.Printf("%-30s %-10s %-50s %-50s\n", host, method, truncate(path, 50), "")
				}
			}
		}
		return
	}

	output, err := json.MarshalIndent(sim, "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize simulation: %v", err)
	}

	if *outputFile != "" {
		err = os.WriteFile(*outputFile, output, 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
	} else {
		fmt.Println(string(output))
	}
}

func isTextContent(mimeType string, allowed []string) bool {
	mimeType = strings.ToLower(mimeType)
	for _, substr := range allowed {
		if strings.Contains(mimeType, substr) {
			return true
		}
	}
	return false
}

func parseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		return &url.URL{}
	}
	return u
}

func convertEntryToPair(entry Entry, sizeLimit int, allowedContentTypes []string) Pair {
	req := entry.Request
	res := entry.Response

	body := res.Content.Text
	if sizeLimit > 0 && len(body) > sizeLimit {
		body = ""
	}

	reqURL := parseURL(req.URL)

	request := Request{
		Method:      []FieldMatcher{{Matcher: "exact", Value: req.Method}},
		Destination: []FieldMatcher{{Matcher: "exact", Value: reqURL.Host}},
		Path:        []FieldMatcher{{Matcher: "exact", Value: reqURL.Path}},
	}

	response := Response{
		Status:  res.Status,
		Body:    body,
		Headers: Header{"Content-Type": []string{res.Content.MimeType}},
	}

	return Pair{
		Request:  request,
		Response: response,
		Labels:   []string{req.Method},
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}
