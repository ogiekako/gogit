package git

import (
	"bytes"
	"fmt"
	"strings"
)

func decodeKVLM(raw string) (map[string][]string, error) {
	m := make(map[string][]string)
	return m, decodeKVLMInner(raw, 0, m)
}

func decodeKVLMInner(raw string, start int, m map[string][]string) error {
	spc := strings.IndexRune(raw[start:], ' ') + start
	nl := strings.IndexRune(raw[start:], '\n') + start

	if spc < start || nl < spc {
		if nl != start {
			return fmt.Errorf("nl != start")
		}
		m[""] = []string{raw[start+1:]}
		return nil
	}

	key := raw[start:spc]

	end := start
	for {
		end = strings.IndexRune(raw[end+1:], '\n') + end + 1
		if raw[end+1] != ' ' {
			break
		}
	}

	value := strings.ReplaceAll(raw[spc+1:end], "\n ", "\n")
	m[key] = append(m[key], value)

	return decodeKVLMInner(raw, end+1, m)
}

func encodeKVLM(m map[string][]string) string {
	var b bytes.Buffer

	for k, v := range m {
		if k == "" {
			continue
		}
		for _, vv := range v {
			fmt.Fprintf(&b, "%s %s\n", k, strings.ReplaceAll(vv, "\n", "\n "))
		}
	}
	fmt.Fprintf(&b, "\n%s", m[""][0])

	return b.String()
}
