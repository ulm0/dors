# Package `markdown`

## Constants

### const [identRx](comment.go#L110)

```go
const (
	// Regexp for Go identifiers
	identRx = `[\pL_][\pL_0-9]*`
)
```

## Variables

### var [`matchRx`](comment.go#L135)

```go
var matchRx = regexp.MustCompile(`(` + urlTitle + `((` + urlRx + `)|(` + localRx + `)))|(` + identRx + `)`)
```

## Functions

### func [`ToMarkdown`](comment.go#L42)

```go
func ToMarkdown(w io.Writer, text string, opts ...Option)
```

ToMarkdown converts comment text to formatted Markdown.
The comment was prepared by DocReader,
so it is known not to have leading, trailing blank lines
nor to have trailing spaces at the end of lines.
The comment markers have already been removed.

Each span of unindented non-blank lines is converted into
a single paragraph. There is one exception to the rule: a span that
consists of a single line, is followed by another paragraph span,
begins with a capital letter, and contains no punctuation
other than parentheses and commas is formatted as a heading.

A span of indented lines is converted into a <pre> block,
with the common indent prefix removed.

URLs in the comment text are converted into links; if the URL also appears
in the words map, the link is taken from the map (if the corresponding map
value is the empty string, the URL is not converted into a link).

### func [`commonPrefix`](comment.go#L280)

```go
func commonPrefix(a, b string) string
```

### func [`diffCharIdx`](comment.go#L470)

```go
func diffCharIdx(line string) int
```

diffCharIdx returns the index of a diff character, given the first line of a code block.

### func [`emphasize`](comment.go#L166)

```go
func emphasize(w io.Writer, line string, words map[string]string, nice bool)
```

Emphasize and escape a line of text for HTML. URLs are converted into links;
if the URL also appears in the words map, the link is taken from the map (if
the corresponding map value is the empty string, the URL is not converted
into a link). Go identifiers that appear in the words map are italicized; if
the corresponding map value is not the empty string, it is considered a URL
and the word is converted into a link.

### func [`heading`](comment.go#L312)

```go
func heading(line string) string
```

heading returns the trimmed line if it passes as a section heading;
otherwise it returns the empty string.

### func [`indentLen`](comment.go#L268)

```go
func indentLen(s string) int
```

### func [`isBlank`](comment.go#L276)

```go
func isBlank(s string) bool
```

### func [`isDiffLine`](comment.go#L488)

```go
func isDiffLine(line string, i int) bool
```

isDiffLine returns if the character at i is a '+' or a '-' sign.

### func [`isValidDiffLine`](comment.go#L480)

```go
func isValidDiffLine(line string, i int) bool
```

isDiffLine returns if this is a valid diff line given a code block line, and the expected index
for the diff character.

### func [`pairedParensPrefixLen`](comment.go#L138)

```go
func pairedParensPrefixLen(s string) int
```

pairedParensPrefixLen returns the length of the longest prefix of s containing paired parentheses.

### func [`unindent`](comment.go#L288)

```go
func unindent(block []string)
```

## Types

### type [`Option`](comment.go#L85)

```go
type Option func(*options)
```

Option is option type for ToMarkdown

#### func [OptNoDiff](comment.go#L96)

```go
func OptNoDiff(noDiffs bool) Option
```

OptNoDiff disables automatic marking of code blocks as diffs.

#### func [OptUseStdlib](comment.go#L100)

```go
func OptUseStdlib(useStdlib bool) Option
```

#### func [OptWords](comment.go#L91)

```go
func OptWords(words map[string]string) Option
```

OptWords sets the list of known words.
Go identifiers that appear in the words map are italicized; if the corresponding
map value is not the empty string, it is considered a URL and the word is converted
into a link.

### type [`block`](comment.go#L370)

```go
type block struct {
	op	op
	lines	[]string

	lang	string	// for opPre, the language of the code block.
}
```

#### func [blocks](comment.go#L377)

```go
func blocks(text string, skipDiffs bool) []block
```

### type [`op`](comment.go#L362)

```go
type op int
```

#### Constants

##### const [`opPara`](comment.go#L364)

```go
const (
	opPara op = iota
)
```

### type [`options`](comment.go#L104)

```go
type options struct {
	words		map[string]string
	noDiffs		bool
	useStdlib	bool	// Use standard library comments parsers introduced in Go 1.19.
}
```
