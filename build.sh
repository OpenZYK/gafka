#!/bin/bash -e

if [[ $1 = "-loc" ]]; then
    find . -name '*.go' | xargs wc -l | sort -n
    exit
fi

VER=0.0.4stable
GOVER=$(go version | cut -d' ' -f3 | cut -d'.' -f2)
# get the git commit
GIT_ID=$(git rev-parse HEAD | cut -c1-7)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)

BUILD_FLAGS=''
if [[ $1 = "-race" ]]; then
    BUILD_FLAGS="$BUILD_FLAGS -race"
fi
if [[ $1 = "-gc" ]]; then
    BUILD_FLAGS="$BUILD_FLAGS -gcflags '-m=1'"
fi

cd cmd/gk/
if [ $GOVER -gt 4 ]; then
    go build $BUILD_FLAGS -tags release -ldflags "-X github.com/funkygao/gafka.Version=$VER -X github.com/funkygao/gafka.BuildId=${GIT_ID}${GIT_DIRTY} -w"
else
    go build $BUILD_FLAGS -tags release -ldflags "-X github.com/funkygao/gafka.Version $VER -X github.com/funkygao/gafka.BuildId ${GIT_ID}${GIT_DIRTY} -w"
fi

if [[ $1 = "-install" ]]; then
    install gk $GOPATH/bin
fi

#---------
# show ver
#---------
./gk -version
