package cli

import (
	"io"
	"regexp"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
)

type inputIter interface {
	gojq.Iter
	io.Closer
	Name() string
}

func makeQueryFriendly(key string) string {
	return strings.ToLower(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(key, " ", "_"),
				"(", "_"),
			")", ""),
	)
}

type procMapIter struct {
	fname   string
	content map[string]interface{}
}

func newProcMapIter(fname string, r io.Reader, parser func(io.Reader) (map[string]interface{}, error)) (inputIter, error) {
	content, err := parser(r)
	if err != nil {
		return nil, err
	}
	return &procMapIter{fname: fname, content: content}, nil
}

func (i *procMapIter) Next() (interface{}, bool) {
	if i.content == nil {
		return nil, false
	}

	result := i.content
	i.content = nil
	return result, true
}

func (i *procMapIter) Close() error {
	i.content = nil
	return nil
}

func (i *procMapIter) Name() string {
	return i.fname
}

func createMapParser(lineParser lineParserFn) func(io.Reader) (map[string]interface{}, error) {
	return func(r io.Reader) (map[string]interface{}, error) {
		b, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		result := make(map[string]interface{})
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			if line == "" {
				break
			}
			key, val, err := lineParser(line)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}

		return result, nil
	}
}

type procArrayIter struct {
	fname   string
	content []interface{}
}

func newProcArrayIter(fname string, r io.Reader, parser func(io.Reader) ([]interface{}, error)) (inputIter, error) {
	content, err := parser(r)
	if err != nil {
		return nil, err
	}
	return &procArrayIter{fname: fname, content: content}, nil
}

func (i *procArrayIter) Next() (interface{}, bool) {
	if i.content == nil {
		return nil, false
	}

	result := i.content
	i.content = nil
	return result, true
}

func (i *procArrayIter) Close() error {
	i.content = nil
	return nil
}

func (i *procArrayIter) Name() string {
	return i.fname
}

type procTableIter struct {
	fname   string
	content []interface{}
}

func newProcTableIter(fname string, r io.Reader, parser tableParserFn) (inputIter, error) {
	content, err := parser(r)
	if err != nil {
		return nil, err
	}
	return &procTableIter{fname: fname, content: content}, nil
}

func (i *procTableIter) Next() (interface{}, bool) {
	if i.content == nil {
		return nil, false
	}

	result := i.content
	i.content = nil
	return result, true
}

func (i *procTableIter) Close() error {
	i.content = nil
	return nil
}

func (i *procTableIter) Name() string {
	return i.fname
}

type tableHeaderParserFn func(rows []string) ([]string, []string, error)

var noTableHeader tableHeaderParserFn = nil

func skipTableHeader(nRows int) tableHeaderParserFn {
	return func(rows []string) ([]string, []string, error) {
		if len(rows) < nRows {
			return nil, nil, errors.Errorf("unable to skip table header: %d rows to be skipped, but only %d rows available", nRows, len(rows))
		}

		return rows[:nRows], rows[nRows:], nil
	}
}

type tableRowParserFn func(header []string, row string) (map[string]interface{}, error)
type tableParserFn func(io.Reader) ([]interface{}, error)
type tableColumnSplitterFn func(row string) ([]string, error)
type tableColumnsParserFn func(header, columns []string) (map[string]interface{}, error)

func createTableParser(headerParser tableHeaderParserFn, rowParser tableRowParserFn) tableParserFn {
	return func(r io.Reader) ([]interface{}, error) {
		lines, err := readAllLines(r)
		if err != nil {
			return nil, err
		}

		var header, remaining []string
		if headerParser != nil && len(lines) > 0 {
			header, remaining, err = headerParser(lines)
			if err != nil {
				return nil, err
			}
			lines = remaining
		}

		var result []interface{}
		for _, line := range lines {
			if line == "" {
				continue
			}
			row, err := rowParser(header, line)
			if err != nil {
				return nil, err
			}
			result = append(result, row)
		}

		return result, nil
	}
}

func createTableRowParser(columnSplitter tableColumnSplitterFn, columnsParser tableColumnsParserFn) tableRowParserFn {
	return func(header []string, row string) (map[string]interface{}, error) {
		columns, err := columnSplitter(row)
		if err != nil {
			return nil, err
		}

		return columnsParser(header, columns)
	}
}

func readAllLines(r io.Reader) ([]string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(b), "\n"), nil
}

type chunkParserFn func(io.Reader) ([]interface{}, error)

func createChunkParser(lineParser lineParserFn) chunkParserFn {
	return func(r io.Reader) ([]interface{}, error) {
		chunks, err := parseAsChunks(r)
		if err != nil {
			return nil, err
		}

		var result []interface{}
		for _, chunk := range chunks {
			m, err := parseChunkLines(chunk, lineParser)
			if err != nil {
				return nil, err
			}

			result = append(result, m)
		}

		return result, nil
	}
}

type lineSplitterFn func(string) (string, string, error)
type valueParserFn func(string, string) (interface{}, error)
type lineParserFn func(string) (string, interface{}, error)

func createLineParser(splitter lineSplitterFn, valueParser valueParserFn) lineParserFn {
	return func(line string) (string, interface{}, error) {
		key, val, err := splitter(line)
		if err != nil {
			return "", "", err
		}

		valParsed, err := valueParser(key, val)
		if err != nil {
			return "", "", err
		}

		if options.OutputQueryFriendly {
			key = makeQueryFriendly(key)
		}
		return key, valParsed, nil
	}
}

type chunk []string

func newChunk() chunk {
	return []string{}
}

func parseAsChunks(r io.Reader) ([]chunk, error) {
	lines, err := readAllLines(r)
	if err != nil {
		return nil, err
	}

	var chunks []chunk
	var curr chunk

	for _, line := range lines {
		if line == "" {
			if len(curr) > 0 {
				chunks = append(chunks, curr)
			}
			curr = newChunk()
			continue
		}

		curr = append(curr, line)
	}

	return chunks, nil
}

func parseChunkLines(chunk chunk, lineParser func(string) (string, interface{}, error)) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, line := range chunk {
		key, value, err := lineParser(line)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func splitLineByColon(line string) (string, string, error) {
	return splitLineBy(line, ":")
}

func splitLineBySpace(line string) (string, string, error) {
	return splitLineBy(line, " ")
}

func splitLineBy(line, sep string) (string, string, error) {
	parts := strings.Split(line, sep)
	if len(parts) < 1 {
		return "", "", errors.New("empty string where colon separated string is expected")
	}

	key := strings.TrimSpace(parts[0])

	val := ""
	if len(parts) >= 2 {
		val = strings.TrimSpace(parts[1])
	}

	return key, val, nil
}

var reNotSpace = regexp.MustCompile(`\S+`)

func splitColumnsBySpace(row string) ([]string, error) {
	return reNotSpace.FindAllString(row, -1), nil
}

var reNotSpaceNorColon = regexp.MustCompile(`[^\s:]+`)

func splitColumnsByColonAndSpace(row string) ([]string, error) {
	return reNotSpaceNorColon.FindAllString(row, -1), nil
}

var reInteger = regexp.MustCompile(`^\d+$`)

func isLikelyInteger(s string) bool {
	return reInteger.MatchString(s)
}
