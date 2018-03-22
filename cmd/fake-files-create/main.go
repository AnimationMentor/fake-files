package main

//go:generate ./embed-content.sh

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

var topLevelDir, inputFilename string
var overwrite bool

func init() {
	flag.StringVar(&topLevelDir, "dir", "", "top level directory to write into")
	flag.StringVar(&inputFilename, "json", "", "json file containing content list")
	flag.BoolVar(&overwrite, "replace-all", false, "replace all files mentioned in the input list, normally we ignore if a file is already there")
}

func usage() {
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {

	flag.Parse()

	if len(flag.Args()) != 0 || topLevelDir == "" || inputFilename == "" {
		usage()
	}

	if _, err := os.Stat(topLevelDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	ll := make(map[string]string, 10000)

	fp, err := os.Open(inputFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer fp.Close()

	dec := json.NewDecoder(fp)

	if err := dec.Decode(&ll); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	go worker(1)
	go worker(2)
	go worker(3)
	go worker(4)
	go worker(5)
	go worker(6)

	for k, v := range ll {
		workChannel <- []string{k, v}
	}
	close(workChannel)
	wg.Wait()
	log.Printf("%d files wrote, %d skipped, %d failed", filesWrote, filesSkipped, filesFailed)
}

var filesWrote, filesFailed, filesSkipped int64
var workChannel = make(chan []string)
var wg = sync.WaitGroup{}

func worker(i int) {
	wg.Add(1)
	defer wg.Done()
	rlog := logrus.WithField("worker", i)
	var myFilesWrote, myFilesSkipped, myFilesFailed int
	for {
		kv := <-workChannel
		if kv == nil {
			break
		}
		w, s, f := makeMock(rlog, kv[0], kv[1])
		myFilesWrote += w
		myFilesFailed += f
		myFilesSkipped += s
	}
	atomic.AddInt64(&filesWrote, int64(myFilesWrote))
	atomic.AddInt64(&filesFailed, int64(myFilesFailed))
	atomic.AddInt64(&filesSkipped, int64(myFilesSkipped))
}

// makeMock creates a mock file of the given type. If the file already exists, it does nothing.
// Return count for files wrote, skipped, and failed.
func makeMock(log *logrus.Entry, filename, contentType string) (int, int, int) {
	// if file already exists then do nothing
	if !overwrite {
		if _, err := os.Stat(filename); err == nil {
			log.Infof("skipping: %s %s", filename, contentType)
			return 0, 1, 0
		}
	}

	// create parent directory if we need to
	parentDir := filepath.Dir(filename)
	if _, err := os.Stat(parentDir); err != nil {
		os.MkdirAll(parentDir, 0775)
	}

	// create the file
	if err := ioutil.WriteFile(filename, getMockContents(contentType), 0664); err != nil {
		log.Errorf("problem writing %s: %v", filename, err)
		return 0, 0, 1
	}

	log.Infof("wrote: %s %s", filename, contentType)
	return 1, 0, 0
}

var contentMap = make(map[string][]byte, 10)

func getMockContents(contentType string) []byte {

	if strings.HasPrefix(contentType, "base64:") {
		b, err := base64.RawStdEncoding.DecodeString(contentType[len("base64:"):])
		if err != nil {
			return []byte(err.Error())
		}
		return b
	}

	if b := contentMap[contentType]; b != nil {
		return b
	}
	return contentMap["text"]
}

func init() {
	contentMap["empty"] = []byte{}
	contentMap["text"] = []byte("Hello, I am mock text file.")
	contentMap["image/png"] = makePNG()
	contentMap["image/jpeg"] = makeJpeg()
}
