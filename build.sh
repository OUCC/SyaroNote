#!/bin/bash

git submodule update
cd $GOPATH/src/github.com/libgit2/git2go
make install
go install
cd $GOPATH/src/github.com/OUCC/syaro
go install
bower update
./install_components.sh
