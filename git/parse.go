package git

import (
	"strings"
	"fmt"
)

func parseKVLM(raw string) (map[string]string, error) {
	m := make(map[string]string)
	return m, parseKVLMInner(raw, 0, m)
}

func parseKVLMInner(raw string, start int, m map[string]string) error {
	spc := strings.IndexRune(raw[start:], ' ') + start
	nl := strings.IndexRune(raw[start:], '\n') + start

	if spc < 0 || nl < spc {
		if nl != start {
			return fmt.Errorf("nl != start")
		}
		m[""] = raw[start + 1:]
		return nil
	}

	key := raw[start:spc]
	if _, ok := m[key]; ok {
		return fmt.Errorf("key %s duplicated", key)
	}

	end := start
	for {
		end = strings.IndexRune(raw[end+1:], '\n') + end + 1
		if raw[end + 1] != ' ' {
			break
		}
	}

	value := strings.ReplaceAll(raw[spc + 1: end], "\n ", "\n")

	m[key] = value

	return parseKVLMInner(raw, end+1, m)
}