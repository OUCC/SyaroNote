#!/bin/bash

# git plugin
cd $GOPATH/src/github.com/libgit2/git2go
git submodule update --init
make install
go install
cd $GOPATH/src/github.com/OUCC/syaro/gitplugin
go build -o build/gitplugin
