// Package inspect produces a compact, LLM-readable outline of a fetched
// document plus candidate selectors/paths, so a config can be authored to
// match on the first try instead of guessing selectors and getting nulls.
package inspect

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

const (
	maxLines     = 120
	maxSample    = 60
	maxJSONDepth = 5
)

// Summarize returns the outline. kind is a hint ("json"/"html"/"xpath"/"xml");
// empty auto-detects from the body.
func Summarize(body []byte, kind string) string {
	trimmed := strings.TrimSpace(string(body))
	k := strings.ToLower(kind)
	if k == "" {
		switch {
		case strings.HasPrefix(trimmed, "{"), strings.HasPrefix(trimmed, "["):
			k = "json"
		case strings.HasPrefix(trimmed, "<"):
			k = "html"
		}
	}
	switch k {
	case "json":
		return summarizeJSON(body)
	case "html", "xpath", "xml":
		return summarizeHTML(body)
	default:
		return "unrecognized content type; first 500 bytes:\n" + truncate(trimmed, 500)
	}
}

func summarizeJSON(body []byte) string {
	if !gjson.ValidBytes(body) {
		return "content is not valid JSON; first 500 bytes:\n" + truncate(strings.TrimSpace(string(body)), 500)
	}
	lines := []string{"response_type: json — gjson paths (use these as field \"path\"):"}
	walkJSON("", gjson.ParseBytes(body), 0, &lines)
	return joinCapped(lines)
}

func walkJSON(prefix string, r gjson.Result, depth int, lines *[]string) {
	if len(*lines) >= maxLines || depth > maxJSONDepth {
		return
	}
	switch {
	case r.IsObject():
		r.ForEach(func(key, val gjson.Result) bool {
			p := joinPath(prefix, key.String())
			*lines = append(*lines, fmt.Sprintf("%s%s: %s", indent(depth), p, describe(val)))
			if val.IsObject() || val.IsArray() {
				walkJSON(p, val, depth+1, lines)
			}
			return len(*lines) < maxLines
		})
	case r.IsArray():
		arr := r.Array()
		rootPath := strings.TrimPrefix(prefix, ".")
		if rootPath == "" {
			rootPath = "@this"
		}
		*lines = append(*lines, fmt.Sprintf("%s%s: [array, %d items] — for array_config use root_path %q; element paths use %q",
			indent(depth), orRoot(prefix), len(arr), rootPath, joinPath(prefix, "0")))
		if len(arr) > 0 {
			walkJSON(joinPath(prefix, "0"), arr[0], depth+1, lines)
		}
	}
}

func summarizeHTML(body []byte) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "unable to parse HTML: " + err.Error()
	}
	var lines []string
	if t := clean(doc.Find("title").First().Text()); t != "" {
		lines = append(lines, "title: "+truncate(t, 100)+`  (css: "title")`)
	}
	if h1 := clean(doc.Find("h1").First().Text()); h1 != "" {
		lines = append(lines, "h1: "+truncate(h1, 100)+`  (css: "h1")`)
	}

	counts := map[string]int{}
	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		cls, ok := s.Attr("class")
		if !ok || strings.TrimSpace(cls) == "" {
			return
		}
		counts[goquery.NodeName(s)+"."+strings.Join(strings.Fields(cls), ".")]++
	})
	type sel struct {
		selector string
		count    int
	}
	var sels []sel
	for s, c := range counts {
		if c >= 3 {
			sels = append(sels, sel{s, c})
		}
	}
	sort.Slice(sels, func(i, j int) bool {
		if sels[i].count != sels[j].count {
			return sels[i].count > sels[j].count
		}
		return sels[i].selector < sels[j].selector
	})
	if len(sels) > 0 {
		lines = append(lines, "", "repeated elements (candidate array_config root_path / list rows):")
		for i, s := range sels {
			if i >= 12 {
				break
			}
			sample := clean(doc.Find(s.selector).First().Text())
			lines = append(lines, fmt.Sprintf("  %s  (%d matches)  e.g. %q", s.selector, s.count, truncate(sample, maxSample)))
		}
	}

	links := doc.Find("a[href]")
	lines = append(lines, "", fmt.Sprintf(`links: %d <a href> (css "a", type "html_attribute" with attribute "href")`, links.Length()))
	links.EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= 5 {
			return false
		}
		href, _ := s.Attr("href")
		lines = append(lines, fmt.Sprintf("  - %q → %s", truncate(clean(s.Text()), 40), truncate(href, 80)))
		return true
	})

	// SPA / empty-shell heuristic: little static text but scripts present means
	// the content is built client-side and won't be in this raw HTML.
	if bodyText := clean(doc.Find("body").Text()); len(bodyText) < 200 {
		if scripts := doc.Find("script").Length(); scripts >= 1 {
			hint := fmt.Sprintf("⚠ likely a client-rendered SPA: only %d chars of static body text but %d <script> tags.", len(bodyText), scripts)
			var mount string
			doc.Find("#root, #app, #__next, [data-reactroot]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
				if id, ok := s.Attr("id"); ok {
					mount = id
				}
				return mount == ""
			})
			if mount != "" {
				hint += fmt.Sprintf(" Found a #%s mount element.", mount)
			}
			hint += " The data is NOT in the raw HTML — re-inspect with render:true (headless browser), or scrape with a browser_config (playwright) connector."
			lines = append([]string{hint, ""}, lines...)
		}
	}

	return joinCapped(lines)
}

func describe(v gjson.Result) string {
	switch {
	case v.IsObject():
		return "object"
	case v.IsArray():
		return fmt.Sprintf("array[%d]", len(v.Array()))
	default:
		return fmt.Sprintf("%s (%s)", strings.ToLower(v.Type.String()), truncate(v.String(), maxSample))
	}
}

func orRoot(prefix string) string {
	if prefix == "" {
		return "(root)"
	}
	return prefix
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func indent(d int) string { return strings.Repeat("  ", d) }

func clean(s string) string { return strings.Join(strings.Fields(s), " ") }

func truncate(s string, n int) string {
	s = clean(s)
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

func joinCapped(lines []string) string {
	if len(lines) > maxLines {
		lines = append(lines[:maxLines:maxLines], "… (truncated)")
	}
	return strings.Join(lines, "\n")
}
