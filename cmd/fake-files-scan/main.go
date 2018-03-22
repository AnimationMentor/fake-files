package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s path\n", os.Args[0])
		return
	}
	skips, types, ll, err := walk(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	b, err := json.MarshalIndent(ll, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	os.Stdout.Write(b)
	fmt.Println("")

	fmt.Fprintf(os.Stderr, "found %d files, skipped %d files, %d types:\n", len(ll), skips, len(types))
	for t, c := range types {
		fmt.Fprintf(os.Stderr, "%s %d\n", t, c)
	}
}

func walk(dir string) (int, map[string]int, map[string]string, error) {
	types := make(map[string]int, 10)
	ll := make(map[string]string, 10000)
	var skips int
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if 0 != (f.Mode() & (os.ModeSymlink | os.ModeDir)) {
			return nil
		}
		t := mime.TypeByExtension(filepath.Ext(path))
		if len(t) == 0 {
			if f.Size() == 0 {
				types["empty"]++
				ll[path] = "empty"
				return nil
			}
			fp, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping: %s (%s)\n", path, err)
				skips++
				return nil
			}
			defer fp.Close()
			b := make([]byte, 512)
			_, err = fp.Read(b)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping: %s (%s)\n", path, err)
				skips++
				return nil
			}

			t = http.DetectContentType(b)
			t = strings.Replace(t, "; charset=utf-8", "", -1)
		}

		if f.Size() > 0 && f.Size() <= 1000 {
			fp, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping: %s (%s)\n", path, err)
				skips++
				return nil
			}
			defer fp.Close()
			b := make([]byte, 1000)
			c, err := fp.Read(b)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping: %s (%s)\n", path, err)
				skips++
				return nil
			}
			ll[path] = "base64:" + base64.RawStdEncoding.EncodeToString(b[0:c])
			types["embedded"]++
		} else {
			types[t]++
			ll[path] = t
		}

		return nil
	})
	return skips, types, ll, err
}
