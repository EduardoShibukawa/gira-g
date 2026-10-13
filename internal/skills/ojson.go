package skills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// setJSONKey sets obj[path[0]]...[path[n-1]] = value in a JSON document by
// splicing bytes: only the owned key's value is replaced (or inserted, with the
// missing parent objects), and every other byte of the file is kept. In a JSON
// file the owned key plays the role of the install markers. If the key already
// holds an equal value the document is returned unchanged.
func setJSONKey(doc []byte, path []string, value any) ([]byte, error) {
	if len(bytes.TrimSpace(doc)) == 0 {
		doc = []byte("{}\n")
	}
	if _, err := decodeOrdered(doc); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	want, err := decodeOrdered(raw)
	if err != nil {
		return nil, err
	}

	start, end, openBrace, depth, err := locate(doc, path)
	if err != nil {
		return nil, err
	}
	if start >= 0 {
		cur, err := decodeOrdered(doc[start:end])
		if err != nil {
			return nil, err
		}
		if equalJSON(cur, want) {
			return doc, nil
		}
		indent := lineIndent(doc, start)
		var buf bytes.Buffer
		if err := encodeOrdered(&buf, want, indent); err != nil {
			return nil, err
		}
		return splice(doc, start, end, buf.Bytes()), nil
	}

	// Insert the missing keys path[depth:] into the object opened at openBrace.
	var v any = want
	for i := len(path) - 1; i > depth; i-- {
		o := &ordered{}
		o.set(path[i], v)
		v = o
	}
	closeAt, empty, err := objectEnd(doc, openBrace)
	if err != nil {
		return nil, err
	}
	parentIndent := lineIndent(doc, openBrace)
	memberIndent := parentIndent + "  "
	if !empty {
		memberIndent = firstMemberIndent(doc, openBrace, memberIndent)
	}
	var buf bytes.Buffer
	kb, _ := json.Marshal(path[depth])
	if !empty {
		buf.WriteByte(',')
	}
	buf.WriteString("\n" + memberIndent)
	buf.Write(kb)
	buf.WriteString(": ")
	if err := encodeOrdered(&buf, v, memberIndent); err != nil {
		return nil, err
	}
	buf.WriteString("\n" + parentIndent)
	// Insert after the last member (trailing whitespace before '}' is replaced).
	insertAt := closeAt
	for insertAt > 0 && (doc[insertAt-1] == ' ' || doc[insertAt-1] == '\n' || doc[insertAt-1] == '\t' || doc[insertAt-1] == '\r') {
		insertAt--
	}
	return splice(doc, insertAt, closeAt, buf.Bytes()), nil
}

func splice(doc []byte, start, end int, with []byte) []byte {
	out := make([]byte, 0, len(doc)-(end-start)+len(with))
	out = append(out, doc[:start]...)
	out = append(out, with...)
	return append(out, doc[end:]...)
}

// locate finds the value of path in doc. If found it returns its byte span;
// otherwise start is -1, openBrace is the offset of the '{' of the deepest object
// on the path that exists, and depth is how many path keys exist.
func locate(doc []byte, path []string) (start, end, openBrace, depth int, err error) {
	dec := json.NewDecoder(bytes.NewReader(doc))
	tok, err := dec.Token()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return 0, 0, 0, 0, fmt.Errorf("top level is not a JSON object")
	}
	openBrace = int(dec.InputOffset()) - 1
	for depth < len(path) {
		found := false
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return 0, 0, 0, 0, err
			}
			valStart := skipColon(doc, int(dec.InputOffset()))
			if kt.(string) != path[depth] {
				if _, err := decodeValue(dec); err != nil {
					return 0, 0, 0, 0, err
				}
				continue
			}
			if depth == len(path)-1 {
				if _, err := decodeValue(dec); err != nil {
					return 0, 0, 0, 0, err
				}
				return valStart, int(dec.InputOffset()), openBrace, depth + 1, nil
			}
			if doc[valStart] != '{' {
				return 0, 0, 0, 0, fmt.Errorf("key %q is not an object", path[depth])
			}
			if _, err := dec.Token(); err != nil { // consume '{'
				return 0, 0, 0, 0, err
			}
			openBrace = valStart
			depth++
			found = true
			break
		}
		if !found {
			return -1, -1, openBrace, depth, nil
		}
	}
	return -1, -1, openBrace, depth, nil
}

func skipColon(doc []byte, i int) int {
	for i < len(doc) && (doc[i] == ' ' || doc[i] == '\t' || doc[i] == '\n' || doc[i] == '\r' || doc[i] == ':') {
		i++
	}
	return i
}

// objectEnd returns the offset of the '}' closing the object opened at open, and
// whether the object has no members.
func objectEnd(doc []byte, open int) (int, bool, error) {
	dec := json.NewDecoder(bytes.NewReader(doc[open:]))
	if _, err := decodeValue(dec); err != nil {
		return 0, false, err
	}
	closeAt := open + int(dec.InputOffset()) - 1
	inner := bytes.TrimSpace(doc[open+1 : closeAt])
	return closeAt, len(inner) == 0, nil
}

// lineIndent returns the leading whitespace of the line containing offset i.
func lineIndent(doc []byte, i int) string {
	ls := bytes.LastIndexByte(doc[:i], '\n') + 1
	j := ls
	for j < len(doc) && (doc[j] == ' ' || doc[j] == '\t') {
		j++
	}
	return string(doc[ls:j])
}

// firstMemberIndent returns the indentation of the first member of the object
// opened at open, or def when it shares a line with the brace.
func firstMemberIndent(doc []byte, open int, def string) string {
	i := open + 1
	for i < len(doc) && (doc[i] == ' ' || doc[i] == '\t' || doc[i] == '\r') {
		i++
	}
	if i >= len(doc) || doc[i] != '\n' {
		return def
	}
	return lineIndent(doc, i+1)
}

func equalJSON(a, b any) bool {
	var ba, bb bytes.Buffer
	_ = encodeOrdered(&ba, sortKeys(a), "")
	_ = encodeOrdered(&bb, sortKeys(b), "")
	return bytes.Equal(ba.Bytes(), bb.Bytes())
}

func sortKeys(v any) any {
	switch t := v.(type) {
	case *ordered:
		o := &ordered{}
		keys := append([]string(nil), t.keys...)
		sort.Strings(keys)
		for _, k := range keys {
			o.set(k, sortKeys(t.vals[k]))
		}
		return o
	case []any:
		out := make([]any, len(t))
		for i, x := range t {
			out[i] = sortKeys(x)
		}
		return out
	}
	return v
}

// ordered is a JSON object that remembers key order.
type ordered struct {
	keys []string
	vals map[string]any
}

func (o *ordered) get(k string) any {
	if o.vals == nil {
		return nil
	}
	return o.vals[k]
}

func (o *ordered) set(k string, v any) {
	if o.vals == nil {
		o.vals = map[string]any{}
	}
	if _, ok := o.vals[k]; !ok {
		o.keys = append(o.keys, k)
	}
	o.vals[k] = v
}

func decodeOrdered(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err == nil {
		return nil, fmt.Errorf("trailing data after JSON value")
	}
	return v, nil
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			o := &ordered{}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				o.set(kt.(string), v)
			}
			_, err := dec.Token()
			return o, err
		case '[':
			arr := []any{}
			for dec.More() {
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			_, err := dec.Token()
			return arr, err
		}
	}
	return tok, nil
}

func encodeOrdered(buf *bytes.Buffer, v any, indent string) error {
	switch t := v.(type) {
	case *ordered:
		if len(t.keys) == 0 {
			buf.WriteString("{}")
			return nil
		}
		buf.WriteString("{\n")
		for i, k := range t.keys {
			kb, _ := json.Marshal(k)
			buf.WriteString(indent + "  ")
			buf.Write(kb)
			buf.WriteString(": ")
			if err := encodeOrdered(buf, t.vals[k], indent+"  "); err != nil {
				return err
			}
			if i < len(t.keys)-1 {
				buf.WriteByte(',')
			}
			buf.WriteByte('\n')
		}
		buf.WriteString(indent + "}")
	case []any:
		if len(t) == 0 {
			buf.WriteString("[]")
			return nil
		}
		buf.WriteString("[\n")
		for i, x := range t {
			buf.WriteString(indent + "  ")
			if err := encodeOrdered(buf, x, indent+"  "); err != nil {
				return err
			}
			if i < len(t)-1 {
				buf.WriteByte(',')
			}
			buf.WriteByte('\n')
		}
		buf.WriteString(indent + "]")
	default:
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(t); err != nil {
			return err
		}
		buf.Truncate(buf.Len() - 1) // Encode appends a newline
	}
	return nil
}

// appendJSONArrayItem appends item to the array at path, creating the array (and
// its parent objects) when missing. If has reports true for an element already in
// the array, the document is returned unchanged. Only the inserted bytes are new.
func appendJSONArrayItem(doc []byte, path []string, item any, has func(any) bool) ([]byte, error) {
	if len(bytes.TrimSpace(doc)) == 0 {
		doc = []byte("{}\n")
	}
	start, end, _, _, err := locate(doc, path)
	if err != nil {
		return nil, err
	}
	if start < 0 {
		return setJSONKey(doc, path, []any{item})
	}
	if doc[start] != '[' {
		return nil, fmt.Errorf("key %q is not an array", path[len(path)-1])
	}
	cur, err := decodeOrdered(doc[start:end])
	if err != nil {
		return nil, err
	}
	for _, el := range cur.([]any) {
		if has(plain(el)) {
			return doc, nil
		}
	}
	raw, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	v, err := decodeOrdered(raw)
	if err != nil {
		return nil, err
	}
	closeAt := end - 1
	inner := bytes.TrimSpace(doc[start+1 : closeAt])
	parentIndent := lineIndent(doc, start)
	memberIndent := parentIndent + "  "
	if len(inner) > 0 {
		memberIndent = firstMemberIndent(doc, start, memberIndent)
	}
	var buf bytes.Buffer
	if len(inner) > 0 {
		buf.WriteByte(',')
	}
	buf.WriteString("\n" + memberIndent)
	if err := encodeOrdered(&buf, v, memberIndent); err != nil {
		return nil, err
	}
	buf.WriteString("\n" + parentIndent)
	insertAt := closeAt
	for insertAt > start+1 && (doc[insertAt-1] == ' ' || doc[insertAt-1] == '\n' || doc[insertAt-1] == '\t' || doc[insertAt-1] == '\r') {
		insertAt--
	}
	return splice(doc, insertAt, closeAt, buf.Bytes()), nil
}

// plain converts an ordered value back to plain maps for inspection.
func plain(v any) any {
	switch t := v.(type) {
	case *ordered:
		m := map[string]any{}
		for _, k := range t.keys {
			m[k] = plain(t.vals[k])
		}
		return m
	case []any:
		out := make([]any, len(t))
		for i, x := range t {
			out[i] = plain(x)
		}
		return out
	}
	return v
}
