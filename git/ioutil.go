package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func writeFile(data []byte, elem ...string) error {
	p := filepath.Join(elem...)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(p, data, 0644)
}
