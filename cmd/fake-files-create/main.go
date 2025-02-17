package main

//go:generate ./embed-content.sh

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/AnimationMentor/fake-files/cmd/fake-files-create/maker"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

var (
	topLevelDir   = "."
	inputFilename string
	overwrite     bool
	dryRun        bool
	workerCount   = 6
)

func init() {
	flag.StringVar(&topLevelDir, "dir", topLevelDir, "directory to write files into")
	flag.StringVar(&inputFilename, "file", "", "file containing content list created by fake-files-scan")
	flag.BoolVar(&overwrite, "replace-all", overwrite, "replace all files mentioned in the input list, normally we ignore if a file is already there")
	flag.BoolVar(&dryRun, "dry-run", dryRun, "don't write anything")
	flag.IntVar(&workerCount, "workers", workerCount, "how many concurrent writers")
}

func usage() {
	flag.PrintDefaults()
	os.Exit(0)
}

type fileEntry struct {
	name        string
	contentType string
	size        int
}

func main() {

	flag.Parse()

	if len(flag.Args()) != 0 || inputFilename == "" {
		usage()
	}

	if _, err := os.Stat(topLevelDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// how many lines are we dealing with
	// get a count so we can make a progress bad
	lineCount, err := countLinesInFile(inputFilename)
	if err != nil {
		log.Fatal(err)
	}

	name := "creating files"
	if dryRun {
		name = "dry run"
	}
	bar = progressbar.Default(int64(lineCount), name)

	fp, err := os.Open(inputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	for i := 1; i <= workerCount; i++ {
		go worker(i)
	}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		fileInfo := strings.Split(scanner.Text(), "\t")
		if len(fileInfo) != 3 {
			log.Fatalf("malformed input: %v", fileInfo)
		}
		size, _ := strconv.Atoi(fileInfo[2])
		workChannel <- &fileEntry{
			name:        filepath.Join(topLevelDir, fileInfo[0]),
			contentType: fileInfo[1],
			size:        size,
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	close(workChannel)
	wg.Wait()
	log.Printf("%d files wrote, %d skipped, %d failed", filesWrote, filesSkipped, filesFailed)
}

var filesWrote, filesFailed, filesSkipped int64
var workChannel = make(chan *fileEntry)
var wg = sync.WaitGroup{}
var bar *progressbar.ProgressBar

func worker(i int) {
	wg.Add(1)
	defer wg.Done()

	rlog := logrus.WithField("worker", i)
	var myFilesWrote, myFilesSkipped, myFilesFailed int
	for {
		fe := <-workChannel
		if fe == nil {
			break
		}
		w, s, f := makeMock(dryRun, rlog, fe.name, fe.contentType, fe.size)
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
func makeMock(dryRun bool, log *logrus.Entry, filename, contentType string, size int) (int, int, int) {
	defer bar.Add(1) // mark one task complete on the progress bar

	// if file already exists then do nothing
	if !overwrite {
		if _, err := os.Stat(filename); err == nil {
			log.Debugf("skipping: %s %s", filename, contentType)
			return 0, 1, 0
		}
	}

	if dryRun {
		log.Debugf("dry run, not creating: %s %s", filename, contentType)
		return 1, 0, 0
	}

	// create parent directory if we need to
	parentDir := filepath.Dir(filename)
	if _, err := os.Stat(parentDir); err != nil {
		os.MkdirAll(parentDir, 0775)
	}

	// create the file
	if err := os.WriteFile(filename, getMockContents(contentType, size), 0664); err != nil {
		log.Errorf("problem writing %s: %v", filename, err)
		return 0, 0, 1
	}

	log.Debugf("wrote: %s %s", filename, contentType)
	return 1, 0, 0
}

var contentMap = make(map[string][]byte, 10)

func getMockContents(contentType string, size int) []byte {

	if size == 0 {
		return []byte{}
	}

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
	contentMap["text"] = []byte("Hello, I am mock text file.")
	contentMap["image/png"] = maker.MakePNG()
	contentMap["image/jpeg"] = maker.MakeJPEG()
}

func countLinesInFile(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	bufferSize := 32 * 1024
	separator := "\n"

	buf := make([]byte, bufferSize)
	var count int

	for {
		n, err := file.Read(buf)
		count += bytes.Count(buf[:n], []byte(separator))

		if err != nil {
			if err == io.EOF {
				return count, nil
			}
			return count, err
		}

	}
}
