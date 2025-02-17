#!/bin/bash

resources=../../resources

cat <<EOF > embedded_content.go
//
// DO NOT EDIT - THIS FILE IS AUTOMATICALLY CREATED BY embed-content.sh
//
package main

import (
    "encoding/base64"
    "log"
)

func unembed(base64Content, logNote string) []byte  {
    b, err := base64.StdEncoding.DecodeString(base64Content)
    if err != nil {
        log.Fatal("failed decoding embedded:", logNote, err)
        return nil
    }
    return b
}

func init() {

EOF

echo -n "const base64SmallMP4 = \`" >> embedded_content.go
base64 ${resources}/small.mp4 >> embedded_content.go
echo "\`" >> embedded_content.go

echo -n "const base64SmallWebm = \`" >> embedded_content.go
base64 ${resources}/small.webm >> embedded_content.go
echo "\`" >> embedded_content.go

echo -n "const base64Gif = \`" >> embedded_content.go
base64 ${resources}/small.gif >> embedded_content.go
echo "\`" >> embedded_content.go

echo -n "const base64Webp = \`" >> embedded_content.go
base64 ${resources}/small.webp >> embedded_content.go
echo "\`" >> embedded_content.go

cat <<EOF >> embedded_content.go

    contentMap["video/mp4"] = unembed(base64SmallMP4, "small.mp4")
    contentMap["video/webm"] = unembed(base64SmallWebm, "small.webm")
    contentMap["image/gif"] = unembed(base64Gif, "small.gif")
    contentMap["image/webp"] = unembed(base64Webp, "small.webp")
}

EOF

