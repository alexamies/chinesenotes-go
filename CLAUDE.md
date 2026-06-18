# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```shell
# Build
go build

# Run all tests with coverage
go test ./... -coverprofile=coverage.out

# Run a single test
go test ./tokenizer/... -run TestTokenize

# View test coverage report
go tool cover -html=coverage.out

# Run the web server (requires CNREADER_HOME for local dictionary loading)
export CNREADER_HOME=.
export CNWEB_HOME=.
./chinesenotes-go
# Server starts at http://localhost:8080

# Build and run without compiling first
go run github.com/alexamies/chinesenotes-go

# Build JavaScript bundle (requires Node.js)
cd web-resources && npm install && npm run-script build && cp dist/cnotes-compiled.* ../web/
```

## Architecture

### Entry Point and HTTP Routing

`cnweb.go` is the main package. The `main()` function registers all HTTP routes and calls `initApp()` to wire up dependencies into a `backends` struct. All handler functions (`findHandler`, `findFullText`, `translationMemory`, etc.) are top-level functions that close over the global `b *backends`.

### Configuration

Two config files are read at startup:
- `config.yaml` — loaded by `config.InitConfig()` using the `CNREADER_HOME` env var (falls back to `.`). Controls dictionary file paths (`DictionaryDir`, `LUFiles`), Firestore index names (`IndexCorpus`, `IndexGen`).
- `webconfig.yaml` — loaded by `config.InitWeb()` using `CNWEB_HOME`. Controls HTML templates (`TemplateDir`), site title, email settings, and notes extraction regex.

### Dictionary Loading

`dictionary.LoadDictFile(appConfig)` loads from TSV files listed in `config.yaml`. If `CNREADER_HOME` is not set, `dictionary.LoadDictURL()` fetches the dictionary from the chinesenotes.com GitHub repo instead (zero-config quickstart mode).

The `dictionary.Dictionary` struct holds two indexes:
- `Wdict map[string]*dicttypes.Word` — forward index keyed by Chinese (simplified or traditional)
- `HeadwordIds map[int]*dicttypes.Word` — lookup by numeric headword ID

### Tokenizer

`tokenizer.DictTokenizer[V]` in `tokenizer/tokenizer.go` tokenizes Chinese text using a greedy algorithm. `Tokenize()` first calls `tokenizer.Segment()` to split text on punctuation and non-Chinese characters, then runs both left-to-right and right-to-left greedy dictionary matching on each Chinese segment, taking whichever produces fewer tokens.

### Document and Full Text Search

Full text search requires an inverted index built by the [cnreader](https://github.com/alexamies/cnreader) tool. The search pipeline:

1. `find.QueryParser.ParseQuery()` tokenizes the query into `TextSegment` slices
2. `find.DocFinder.FindDocuments()` (advanced=true) queries term and bigram frequencies via `find.TermFreqDocFinder`
3. BM25 scores for unigrams and bigrams are combined with title similarity via a logistic regression model (`find.WEIGHT` in `find/document_finder.go`)
4. `fulltext.GetMatches()` retrieves text snippets from corpus files to populate `MatchDetails`

The `TermFreqDocFinder` is implemented by `termfreq.fsDocFinder`, which reads from **Firestore** collections named after `IndexCorpus`/`IndexGen` from `config.yaml`. This means **full text search requires Firestore** (`PROJECT_ID` env var) and the index must be loaded.

Corpus text for snippet retrieval comes from either GCS (`TEXT_BUCKET` env var → `fulltext.GCSLoader`) or local filesystem (`CORPUS_DIR` env var → `fulltext.LocalTextLoader`, defaulting to `../corpus`).

### Storage Backends

The app has multiple optional backends controlled by env vars:
- **Firestore** (`PROJECT_ID`): translation memory, full text search index, substring dictionary index, user authentication
- **MySQL/MariaDB** (`DBUSER`, `DBPASSWORD`, `DATABASE`, `DBHOST`): alternative backend; DDL in `data/chinesenotes.ddl`
- **GCS** (`TEXT_BUCKET`): corpus plain text files for full text search snippets
- **File system**: default fallback for dictionary and document title index (`index/documents.tsv`, `index/keyword_index.json`)

### Authentication

Enabled only when `PROTECTED=true`. Uses Firestore to store sessions and SHA-256 hashed passwords. Password reset emails go via SendGrid (`SENDGRID_API_KEY`).

## HTTP API Endpoints

| Endpoint | Description |
|---|---|
| `GET /find/?query=<chinese_or_english>` | Dictionary lookup + document title search; returns JSON or HTML |
| `GET /findadvanced/?query=<text>` | Full text search of corpus body; requires Firestore + index |
| `GET /findsubstring?query=<term>` | Substring dictionary search; requires Firestore |
| `GET /findtm?query=<chinese>` | Translation memory search; requires Firestore |
| `GET /healthcheck` | Health probe |
| `GET /words/<id>.html` | Word detail page by headword ID |

### Full Text Search Example

```bash
curl "http://localhost:8080/findadvanced/?query=%E3%80%8A%E6%B0%B4%E6%BB%B8%E5%82%B3%E3%80%8B%E8%80%85%EF%BC%8C%E7%99%BC%E6%86%A4%E4%B9%8B%E6%89%80%E4%BD%9C%E4%B9%9F%E3%80%82"
```

Which is the URL-encoded form of:
```bash
curl "http://localhost:8080/findadvanced/?query=《水滸傳》者，發憤之所作也。"
```

The response is a JSON `QueryResults` object. Each `Document` in the `Documents` array includes a `MatchDetails` field with `Snippet` (surrounding context) and `LongestMatch` (the longest query substring found in that document).

Full text search requires the Firestore index to be populated and either GCS (`TEXT_BUCKET`) or local (`CORPUS_DIR`) corpus text files to be accessible.

## Key Environment Variables

| Variable | Purpose |
|---|---|
| `CNREADER_HOME` | Project home for dictionary/corpus files; omit to load dict from URL |
| `CNWEB_HOME` | Location of `webconfig.yaml` |
| `PROJECT_ID` | GCP project ID; enables Firestore backend |
| `TEXT_BUCKET` | GCS bucket for corpus plain text files |
| `CORPUS_DIR` | Local fallback directory for corpus text |
| `PROTECTED` | Set to `true` to enable password protection |
| `SITEDOMAIN` | Domain used for session cookies |
| `DEEPL_AUTH_KEY` | DeepL machine translation API key |
| `TRANSLATION_GLOSSARY` | Google Translate glossary name |
| `SENDGRID_API_KEY` | For password reset email delivery |
