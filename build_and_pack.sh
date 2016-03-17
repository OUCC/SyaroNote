#!/bin/bash

copy_files () {
    cp -a README.md LICENSE INSTALL.md build/public build/template syaronote/
}

build () {
    bin=syaronote/syaro
    if [ $1 == "windows" ]; then
        bin=syaronote/syaro.exe
    fi
    arch_=amd64
    if [ $2 == "32" ]; then
        arch_="386"
    fi
    GOOS=$1 GOARCH=${arch_} go build -o $bin -ldflags "-X main.version=$(git describe)" github.com/OUCC/SyaroNote/syaro
    if [ $1 == "linux" ]; then
        GOOS=$1 GOARCH=${arch_} go build -o syaronote/gitplugin -ldflags "-X main.version=$(git describe)" github.com/OUCC/SyaroNote/gitplugin
    fi
}

compress () {
    if [ "$1" == "windows" ]; then
        zip -r ,/syaronote-${1}-${2}bit-$(git describe).zip syaronote
    else
        tar czf ,/syaronote-${1}-${2}bit-$(git describe).tar.gz syaronote
    fi
}

build_and_compress () {
    os=$1
    arch=$2
    echo "${os} ${arch}bit"
    mkdir syaronote
    echo building...
    build $os $arch
    echo copying...
    copy_files
    echo compressing...
    compress $os $arch
    rm -rf syaronote
    echo "done (${os} ${arch}bit)"
}

# editorjs
cd editorjs
npm install
npm run deploy

cd ..
gulp copy

if [ $# == 2 ]; then
    build_and_compress $1 $2
    exit 0
fi

for os in linux darwin windows; do
    build_and_compress $os 64
done

for os in linux windows; do
    build_and_compress $os 32
done
