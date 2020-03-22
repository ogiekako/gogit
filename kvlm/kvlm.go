package kvlm

import (
	"bytes"
	"fmt"
	"strings"
)

type KVLM struct {
	keys []string
	m    map[string][]string
}

func New() *KVLM {
	return &KVLM{m: make(map[string][]string)}
}

func (kv *KVLM) Append(k, v string) {
	if _, ok := kv.m[k]; !ok {
		kv.keys = append(kv.keys, k)
	}
	kv.m[k] = append(kv.m[k], v)
}

func (kv *KVLM) Get(k string) []string {
	return kv.m[k]
}

func Decode(raw string) (*KVLM, error) {
	kv := New()
	return kv, decodeInner(raw, 0, kv)
}

func decodeInner(raw string, start int, kv *KVLM) error {
	spc := strings.IndexRune(raw[start:], ' ') + start
	nl := strings.IndexRune(raw[start:], '\n') + start

	if spc < start || nl < spc {
		if nl != start {
			return fmt.Errorf("nl != start")
		}
		kv.Append("", raw[start+1:])
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
	kv.Append(key, value)

	return decodeInner(raw, end+1, kv)
}

func Encode(kv *KVLM) string {
	var b bytes.Buffer

	for _, k := range kv.keys {
		if k == "" {
			continue
		}
		for _, vv := range kv.m[k] {
			fmt.Fprintf(&b, "%s %s\n", k, strings.ReplaceAll(vv, "\n", "\n "))
		}
	}
	fmt.Fprintf(&b, "\n%s", kv.m[""][0])

	return b.String()
}
