#!/bin/bash

cd $GOPATH/src/github.com/libgit2/git2go
git submodule update --init
make install
go install
cd $GOPATH/src/github.com/OUCC/syaro
go install
bower update
git submodule update --init
./install_components.sh
