
# fake-files

This is a pair of tools which were written to populate a QA environment using content similar to a file set in our production
environment.

- We needed all files from production to exist in QA.
- We needed the files to be mostly of the same type as in production.
- We did not need or want them to be identical to those in production in part for confidentiality and in part it was just too massive a file set to copy.

`fake-files-scan` collects information about existing files. It walks a directory gathering a names of files and attempting to detect their content type. It emits a tab separated file with lines containing three fields: name, content type and file size. Files under a certain size are base64 encoded and embedded in the content type field.

`fake-files-create` reads this output and creates mock file content.

```bash
$ fake-files-scan -start-dir testdata > out-files
INFO[0000] starting from /home/reed/go/src/github.com/AnimationMentor/fake-files/testdata
INFO[0000] found 9 files, skipped 0 files, 6 types
INFO[0000] image/gif 1
INFO[0000] image/png 1
INFO[0000] video/mp4 1
INFO[0000] video/webm 1
INFO[0000] embedded 4
INFO[0000] image/jpeg 1
$ cat out-files
samples/embeded.txt base64:TnVsbGEgcG9ydHRpdG9yIGFjY3Vtc2FuIHRpbmNpZHVudC4gVmVzdGlidWx1bSBhbnRlIGlwc3VtIHByaW1pcyBpbiBmYXVjaWJ1cyBvcmNpIGx1Y3R1cyBldCB1bHRyaWNlcyBwb3N1ZXJlIGN1YmlsaWEgQ3VyYWUuCg 124
samples/empty.txt   base64: 0
samples/hello       base64:aGVsbG8K 6
samples/image.gif   image/gif       2586
samples/image.jpeg  image/jpeg      5656
samples/image.png   image/png       2017
samples/plain.txt   base64:SGVsbG8sIEkgYW0gbW9jayB0ZXh0IGZpbGUu 27
samples/video.mp4   video/mp4       383631
samples/video.webm  video/webm      229455
$
```

## Build

These tools are written in Go. There is a makefile.

Running `make` should build the tools and install them to `~/go/bin/`

Running `make full` will run the tests, vet and lint.

The `generate` step embeds mock content into the tool. It requires `bash` and `base64`.

## Performance Tweaks

`fake-files-scan` is intentionally not written to take avantage of concurrency. I wanted to avoid placing any load on our production
filers. (This should be added as an option.)

`fake-files-create` uses 6 concurrent workers. A different number may work better for you. Beating up our QA filers is OK.
