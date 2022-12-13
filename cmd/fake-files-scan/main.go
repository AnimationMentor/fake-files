package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	startDir        = "."
	embedSize int64 = 1000

	startDirLen int
)

func init() {
	flag.StringVar(&startDir, "start-dir", startDir, "where to start scanning")
	flag.Int64Var(&embedSize, "embed-size", embedSize, "files of this size or smaller are embedded (set to -1 to avoid catching empty files)")
}

func main() {
	flag.Parse()

	startDir, _ = filepath.Abs(startDir)

	startDirLen = len(startDir)
	if !strings.HasSuffix(startDir, "/") {
		startDirLen++
	}

	log.Printf("starting from %v", startDir)

	skips, types, count, err := walk(startDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("found %d files, skipped %d files, %d types", count, skips, len(types))
	for t, c := range types {
		log.Printf("%s %d", t, c)
	}
}

func report(path, fileType string, size int64) {
	fmt.Printf("%v\t%v\t%v\n", path[startDirLen:], fileType, size)
}

func walk(dir string) (int, map[string]int, int, error) {
	types := make(map[string]int, 10)
	var skips, count int
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Error(err)
			return nil
		}

		if f.Mode()&(os.ModeSymlink|os.ModeDir) != 0 {
			// skip dirs and symlinks
			return nil
		}

		t := mime.TypeByExtension(filepath.Ext(path))

		if len(t) == 0 && f.Size() > 0 {
			fp, err := os.Open(path)
			if err != nil {
				log.Warnf("skipping: %s (%s)", path, err)
				skips++
				return nil
			}
			defer fp.Close()
			b := make([]byte, 512)
			if _, err = fp.Read(b); err != nil {
				log.Warnf("skipping: %s (%s)", path, err)
				skips++
				return nil
			}

			t = http.DetectContentType(b)
		}

		t = strings.ReplaceAll(t, "; charset=utf-8", "")

		if f.Size() <= embedSize {
			if f.Size() == 0 {
				count++
				report(path, "base64:", 0)
				types["embedded"]++
				return nil
			}
			fp, err := os.Open(path)
			if err != nil {
				log.Warnf("skipping: %s (%s)", path, err)
				skips++
				return nil
			}
			defer fp.Close()
			b := make([]byte, embedSize)
			c, err := fp.Read(b)
			if err != nil {
				log.Warnf("skipping: %s (%s)", path, err)
				skips++
				return nil
			}
			count++
			report(path, "base64:"+base64.RawStdEncoding.EncodeToString(b[0:c]), f.Size())
			types["embedded"]++
		} else {
			types[t]++
			count++
			report(path, t, f.Size())
		}

		return nil
	})
	return skips, types, count, err
}
