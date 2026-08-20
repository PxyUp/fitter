package inspect_test

import (
	"strings"
	"testing"

	"github.com/PxyUp/fitter/pkg/inspect"
	"github.com/stretchr/testify/assert"
)

func TestSummarizeJSON(t *testing.T) {
	body := []byte(`{"repo":"fitter","stars":400,"topics":["mcp","scraping"],"owner":{"login":"PxyUp"}}`)
	out := inspect.Summarize(body, "")
	assert.Contains(t, out, "response_type: json")
	assert.Contains(t, out, "repo: string (fitter)")
	assert.Contains(t, out, "stars: number (400)")
	assert.Contains(t, out, "owner.login: string (PxyUp)")
	assert.Contains(t, out, "topics: [array, 2 items]")
}

func TestSummarizeJSONRootArray(t *testing.T) {
	out := inspect.Summarize([]byte(`[{"name":"a"},{"name":"b"}]`), "json")
	assert.Contains(t, out, `root_path "@this"`)
	assert.Contains(t, out, "0.name: string (a)")
}

func TestSummarizeHTML(t *testing.T) {
	body := []byte(`<html><head><title>Hello</title></head><body>
		<h1>Top News</h1>
		<div class="item row"><a href="/a">Story A</a></div>
		<div class="item row"><a href="/b">Story B</a></div>
		<div class="item row"><a href="/c">Story C</a></div>
	</body></html>`)
	out := inspect.Summarize(body, "")
	assert.Contains(t, out, "title: Hello")
	assert.Contains(t, out, "h1: Top News")
	assert.Contains(t, out, "div.item.row  (3 matches)")
	assert.True(t, strings.Contains(out, "links: 3"))
}

func TestSummarizeInvalidJSON(t *testing.T) {
	out := inspect.Summarize([]byte(`not json at all`), "json")
	assert.Contains(t, out, "not valid JSON")
}

func TestSummarizeDetectsSPA(t *testing.T) {
	body := []byte(`<html><head><title>App</title></head><body><div id="root"></div>` +
		`<script src="/bundle.js"></script><script>window.__STATE__={}</script></body></html>`)
	out := inspect.Summarize(body, "")
	assert.Contains(t, out, "likely a client-rendered SPA")
	assert.Contains(t, out, "#root mount element")
	assert.Contains(t, out, "render:true")
}

func TestSummarizeContentHTMLNotFlaggedSPA(t *testing.T) {
	// a content-rich page should NOT be flagged as a SPA even with a script tag
	body := []byte(`<html><body><script>var x=1</script><article>` +
		strings.Repeat("Real article content that a browser would have rendered. ", 10) +
		`</article></body></html>`)
	out := inspect.Summarize(body, "")
	assert.NotContains(t, out, "likely a client-rendered SPA")
}
