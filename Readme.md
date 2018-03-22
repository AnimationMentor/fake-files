
# fake-files

This is a pair of tools which were written to populate a QA environment using content similar to a file set in our production
environment.

- We needed all files from production to exist in QA.
- We needed the files to be mostly of the same type as in production.
- We did not need or want them to be identical to those in production in part for confidentiality and in part it was just too massive a file set to copy.

`fake-files-scan` takes a directory as it's only argument. It walks the directory gathering a list of files and attempting to detect their content type. It emits a json file which looks like:

```
{
    "tmp/test/image.png": "image/png",
    "tmp/test/image.gif": "image/gif",
    "tmp/test/image.jpeg": "image/jpeg",
    "tmp/test/video.mp4": "video/mp4",
    "tmp/test/video.webm": "video/webm",
    "tmp/test/plain.txt": "text/plain",
    "tmp/test/empty.txt": "empty",
    "tmp/test/embeded.txt": "base64:TnVsbGEgcG9ydHRpdG9yIGFjY3Vtc2FuIHRpbmNpZHVudC4gVmVzdGlidWx1bSBhbnRlIGlwc3VtIHByaW1pcyBpbiBmYXVjaWJ1cyBvcmNpIGx1Y3R1cyBldCB1bHRyaWNlcyBwb3N1ZXJlIGN1YmlsaWEgQ3VyYWUuCg"
}
```


`fake-files-create` reads this output and creates mock file content.

```
$ fake-files-create -dir . -json test1.json 
INFO[0000] wrote: tmp/test/image.gif image/gif           worker=6
INFO[0000] wrote: tmp/test/plain.txt text/plain          worker=4
INFO[0000] wrote: tmp/test/empty.txt empty               worker=2
INFO[0000] wrote: tmp/test/image.jpeg image/jpeg         worker=1
INFO[0000] wrote: tmp/test/video.mp4 video/mp4           worker=5
INFO[0000] wrote: tmp/test/image.png image/png           worker=4
INFO[0000] wrote: tmp/test/embeded.txt base64:TnVsbGEgcG9ydHRpdG9yIGFjY3Vtc2FuIHRpbmNpZHVudC4gVmVzdGlidWx1bSBhbnRlIGlwc3VtIHByaW1pcyBpbiBmYXVjaWJ1cyBvcmNpIGx1Y3R1cyBldCB1bHRyaWNlcyBwb3N1ZXJlIGN1YmlsaWEgQ3VyYWUuCg  worker=6
INFO[0000] wrote: tmp/test/video.webm video/webm         worker=3
2018/03/22 17:49:07 8 files wrote, 0 skipped, 0 failed
$
```

## Build

These tools are written in Go. There is a makefile.

Running `make` should build the tools and install them to `~/go/bin/`

Running `make full` will run the tests, vet and lint.

The `generate` step embeds mock content into the tool. It requires `bash` and `base64`.

## Performance Tweaks

`fake-files-scan` has not been made to use any meaningful concurrency. I wanted to avoid placing any load on our production
filers. (Should add as an option.)

`fake-files-create` uses 6 concurrent workers. A different number may work better for you.

## Future Work

There's a lot of possible refinements.

This shouldn't use a json file - I was in a hurry and not thinking. It's not really very helpful in this case and causes it not to scale because I'm reading the whole thing into memory before the creates. Something like a tab separated filename/content type pair per line would make more sense. It's not too horrible as it is though.
