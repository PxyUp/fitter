package main

const configReference = `# Fitter config reference (condensed)

A config for fitter_run/fitter_run_file is a "CliItem" object (JSON or YAML):

{
  "item": { "connector_config": {...}, "model": {...} },   // required
  "limits": {...},                                          // optional
  "references": {...}                                       // optional
}

## item.connector_config — where the data comes from

{
  "response_type": "json" | "HTML" | "XML" | "xpath",  // required: how the fetched body is parsed
  "url": "https://example.com",                        // used by server/browser connectors; supports placeholders
  "attempts": 3,                                       // optional retries
  "null_on_error": false,                              // return null instead of failing

  // exactly ONE of the following connector configs:
  "server_config":  { "method": "GET", "headers": {"Authorization": "Bearer {{{RefName=Token}}}"}, "timeout": 30, "body": "", "json_raw_body": {}, "proxy": {"server": "http://host:3128", "username": "", "password": ""}, "oauth2": {"token_url": "https://.../token", "grant_type": "client_credentials"|"refresh_token", "client_id": "{{{FromEnv=ID}}}", "client_secret": "", "scopes": [], "refresh_token": "", "endpoint_params": {}, "auth_style": ""|"header"|"params", "token_file": "~/.fitter/tokens/name.json"} },   // oauth2: token fetched/refreshed/cached automatically, sent as Authorization header; token_file persists rotated refresh tokens between runs (create it once with the "fitter_cli auth" command)
  "static_config":  { "value": "string value (can be html/json)", "raw": {"any": "json"} },
  "file_config":    { "path": "/path/to/file", "use_formatting": false },
  "int_sequence_config": { "start": 0, "end": 10, "step": 1 },   // [start, end) like range(); good for pagination
  "reference_config": { "name": "MyRef" },             // read prefetched value from top-level references
  "plugin_connector_config": { "name": "my_plugin", "config": {...} },  // requires FITTER_PLUGINS env on the MCP server
  "browser_config": {                                  // headless browser (JS-rendered pages), one of:
    "playwright": { "browser": "Chromium"|"FireFox"|"WebKit", "install": true, "timeout": 30, "wait": 30, "type_of_wait": "load"|"domcontentloaded"|"networkidle"|"commit", "stealth": false, "pre_run_script": "", "post_run_script": "", "storage_state_file": "", "indexed_db": false, "proxy": {...} },   // timeout/wait in SECONDS; pre_run_script = init script injected before page scripts run (no DOM access), post_run_script = evaluated after load before reading content; storage_state_file = path to playwright storage state json (cookies+localStorage) for logged-in sessions, loaded before navigation and written back after each run (create once with the "fitter_cli browser-login" command)
    "chromium":   { "path": "/path/to/chromium", "timeout": 30, "wait": 10000, "flags": [] },   // timeout sec, wait MILLISECONDS
    "docker":     { "image": "docker.io/zenika/alpine-chrome:with-node", "entry_point": "chromium-browser", "timeout": 30, "wait": 10000, "flags": [], "purge": true, "no_pull": false, "pull_timeout": 60 }   // timeout sec, wait msec
  }
}

response_type selects the selector language used in field "path" values:
- "json"  -> gjson paths, e.g. "data.items.0.name", "products.#.title", "@this|@keys"
- "HTML"  -> goquery/CSS selectors, e.g. "div.article > a"
- "xpath" -> XPath, e.g. "//channel/title/text()" (also usable for HTML pages)
- "XML"   -> xmlquery/XPath

## item.model — what to extract

Exactly one of:
- "base_field":    extract a single scalar
- "object_config": build an object
- "array_config":  build an array
Plus "is_array": true to force array output (useful with model fields).

object_config:
{ "fields": { "<name>": <Field>, ... }, "condition": "" }   // or "field": <BaseField> (for arrays of scalars), or nested "array_config"

Field:
{
  "base_field": <BaseField>,
  "object_config": <ObjectConfig>,               // nested object
  "array_config": <ArrayConfig>,                 // nested array
  "first_of": [<Field>, ...]                     // first non-empty resolved field wins
}

BaseField:
{
  "type": "string"|"int"|"int64"|"float"|"float64"|"boolean"|"html"|"raw_string"|"null"|"array"|"object",
  "path": "<selector in the response_type language; relative when inside an array item>",
  "html_attribute": "href",                      // HTML parsing only: take attribute instead of text
  "condition": "fRes > 0",                       // optional expr-lang check on the EXTRACTED value; false = field omitted (no null), generated work skipped
  "generated": <GeneratedFieldConfig>,           // computed instead of extracted
  "first_of": [<BaseField>, ...]
}
Notes: "string" is trimmed and special chars are replaced — use "raw_string" for the plain value.
"html" type only works when the connector returns HTML.

ArrayConfig:
{
  "root_path": "<selector of the repeating element>",
  "item_config": <ObjectConfig>,                 // model of each element (paths relative to root_path)
  "length_limit": 10,
  "reverse": false,
  "condition": "",                               // optional: false = whole array omitted from the parent
  "item_condition": "fSrc.in_stock && fRes.price > 0",  // optional filter over every BUILT item (fRes = item, fSrc = source element, fIndex = index); false items are dropped. Not applied to static_array
  "static_array": { "length": 3, "items": { "0": <Field>, ... } }   // fixed-length array, key = index
}

GeneratedFieldConfig (pick one):
{
  "static":     { "type": "int", "value": "65" } or { "type": "array", "raw": [65, 45] },
  "uuid":       { "regexp": "" },                 // random UUID v4, optional regexp to take a part
  "formatted":  { "template": "https://news.ycombinator.com/item?id={PL}" },   // {PL} = current field value
  "calculated": { "type": "boolean", "expression": "fRes > 500" },             // expr-lang expression
  "model":      { "connector_config": {...}, "model": {...}, "type": "object", "path": "", "expression": "" },  // sub-request per value; {PL} in its url injects the parent value
  "plugin":     { "name": "...", "config": {...} },
  "file":       { "url": "https://host{PL}", "file_name": "", "path": "/dir", "config": { "method": "GET" } },  // download file; result = local path
  "file_storage": { "content": "{PL}\n", "file_name": "out.csv", "path": "/tmp", "append": true }               // write value to local file
}

calculated/expression predefined values (expr-lang):
- fRes     — parsed value of the base field (typed)
- fResJson — JSON string of the value;  fResRaw — value as bytes
- fIndex   — index in the parent array (if any)
- FNull / FNil / isNull(v) / FNewLine

Conditional fields: base_field/object_config/array_config accept "condition",
array_config also "item_condition". Anything except true (incl. errors) OMITS
the field/item from the output — no null is produced. base_field conditions see
the extracted value in fRes and the node it was resolved from (siblings
included) in fSrc; object/array conditions run before resolution against the
source node; item_condition sees each built item (fRes), the source element it
was built from (fSrc — filter on attributes without extracting them) and its
index (fIndex) — use it for declarative array filtering.

## Placeholders (usable in url, headers, body, templates, file paths, values)

- {PL}                          — current/parent field value
- {INDEX} / {HUMAN_INDEX}       — index in parent array (0-based / 1-based)
- {{{json_path}}}               — value from the propagated object/array field, e.g. {{{latitude}}}
- {{{RefName=SomeName}}}        — value of a reference; {{{RefName=SomeName json.path}}} extracts by path
- {{{FromInput=.}}}             — the "input" argument of the tool call; {{{FromInput=json.path}}} extracts by path
- {{{FromEnv=ENV_KEY}}}         — environment variable
- {{{FromExp=fRes + 5}}}        — expr-lang expression
- {{{FromFile=./file.txt}}}     — file content (may itself contain placeholders)
- {{{FromURL=http://host}}}     — response body of a GET request

## references (top level, optional)

Named values prefetched before processing; use via reference_config connector or {{{RefName=Name}}}.
Good for auth tokens shared across requests.

"references": {
  "Token": {
    "connector_config": { "response_type": "json", "url": "https://auth.example.com/token", "server_config": { "method": "POST" } },
    "model": { "base_field": { "type": "string", "path": "access_token" } },
    "type": "string",
    "expire": 3600     // sec; omitted = cached forever, 0 = refetch every time
  }
}

## limits (top level, optional)

{ "host_request_limiter": {"example.com": 5}, "chromium_instance": 1, "docker_containers": 1, "playwright_instance": 1 }

## item.notifier_config (optional) — push the result somewhere after parsing

The parse result is still returned by the tool; the notifier additionally delivers it.

{
  "expression": "len(fResRaw) > 0",   // optional expr-lang condition: notify only when true (fRes* = parse result)
  "template": "",                     // optional template applied to the result before sending (placeholders allowed)
  "send_array_by_item": false,        // send each array element as a separate notification
  "force": false,                     // notify even on parse error

  // exactly ONE of:
  "http":         { "url": "https://hooks.example.com", "method": "POST", "headers": {}, "timeout": 30 },
  "telegram_bot": { "token": "{{{FromEnv=TG_TOKEN}}}", "users_id": [123], "pretty": true, "only_msg": false },
  "redis":        { "addr": "localhost:6379", "password": "", "db": 0, "channel": "fitter" },
  "file":         { "content": "{PL}\n", "file_name": "out.log", "path": "/tmp", "append": true },
  "console":      { "only_result": true }    // prints to the server's stderr, visible in MCP client logs
}

## NOT available via MCP (service mode only)

"trigger_config" (scheduler/http triggers) and "http_server" are only used by the long-running
fitter service binary; in MCP tool calls they are ignored. Each tool call is one-shot: it runs
the config once and returns the result.

## Examples

1) JSON API -> array of objects:
{
  "item": {
    "connector_config": { "response_type": "json", "url": "https://dummyjson.com/products", "server_config": { "method": "GET" } },
    "model": {
      "array_config": {
        "root_path": "products",
        "length_limit": 5,
        "item_config": { "fields": {
          "title": { "base_field": { "type": "string", "path": "title" } },
          "price": { "base_field": { "type": "float", "path": "price" } }
        } }
      }
    }
  }
}

2) Nested API calls (fetch details per list item via generated model; {PL} = current item value):
{
  "item": {
    "connector_config": { "response_type": "json", "url": "https://hacker-news.firebaseio.com/v0/topstories.json", "server_config": { "method": "GET" } },
    "model": { "array_config": {
      "root_path": "@this",
      "length_limit": 3,
      "item_config": { "fields": {
        "id": { "base_field": { "type": "int" } },
        "story": { "base_field": { "type": "int", "generated": { "model": {
          "type": "object",
          "connector_config": { "response_type": "json", "url": "https://hacker-news.firebaseio.com/v0/item/{PL}.json", "server_config": { "method": "GET" } },
          "model": { "object_config": { "fields": {
            "title": { "base_field": { "type": "string", "path": "title" } },
            "score": { "base_field": { "type": "int", "path": "score" } }
          } } }
        } } } }
      } }
    } }
  }
}

3) RSS/XML via xpath -> object with nested array:
{
  "item": {
    "connector_config": { "response_type": "xpath", "url": "https://news.google.com/rss", "server_config": { "method": "GET" } },
    "model": { "object_config": { "fields": {
      "feed_title": { "base_field": { "type": "string", "path": "//channel/title/text()" } },
      "articles": { "array_config": {
        "root_path": "//item", "length_limit": 10,
        "item_config": { "fields": {
          "title": { "base_field": { "type": "string", "path": "title/text()" } },
          "link":  { "base_field": { "type": "string", "path": "link/text()" } }
        } }
      } }
    } } }
  }
}

4) HTML page with CSS selectors and attribute extraction:
{
  "item": {
    "connector_config": { "response_type": "HTML", "url": "https://news.ycombinator.com", "server_config": { "method": "GET" } },
    "model": { "array_config": {
      "root_path": "tr.athing",
      "length_limit": 10,
      "item_config": { "fields": {
        "title": { "base_field": { "type": "string", "path": "td.title > span.titleline > a" } },
        "url":   { "base_field": { "type": "string", "path": "td.title > span.titleline > a", "html_attribute": "href" } }
      } }
    } }
  }
}

5) JS-rendered page via Playwright (timeout/wait in seconds):
{
  "item": {
    "connector_config": {
      "response_type": "HTML", "url": "https://example.com/spa",
      "browser_config": { "playwright": { "browser": "Chromium", "install": true, "timeout": 30, "wait": 30, "type_of_wait": "networkidle" } }
    },
    "model": { "object_config": { "fields": {
      "heading": { "base_field": { "type": "string", "path": "h1" } }
    } } }
  },
  "limits": { "playwright_instance": 1 }
}

6) Using tool input ({{{FromInput=.}}}) and pagination via int_sequence:
{
  "item": {
    "connector_config": { "response_type": "json", "int_sequence_config": { "start": 1, "end": 4 } },
    "model": { "array_config": {
      "root_path": "@this",
      "item_config": { "fields": {
        "page": { "base_field": { "type": "int" } },
        "data": { "base_field": { "type": "int", "generated": { "model": {
          "type": "object",
          "connector_config": { "response_type": "json", "url": "https://api.example.com/{{{FromInput=.}}}/items?page={PL}", "server_config": { "method": "GET" } },
          "model": { "base_field": { "type": "object", "path": "@this" } }
        } } } }
      } }
    } }
  }
}
`
