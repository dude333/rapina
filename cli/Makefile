SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=rapina
WINBINARY=rapina.exe
OSXBINARY=rapina-osx

VERSION=`git describe --tags --always`
BUILD_TIME=`date +%F`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD_TIME}"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) $(wildcard ../*.go) $(wildcard ../parsers/*.go) $(wildcard ../reports/*.go)
	go get -v -d ../...
	go build ${LDFLAGS} -o $(BINARY)

win: $(SOURCES)
	go get -v -d ../...
	GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ CGO_LDFLAGS="-lssp -w" go build ${LDFLAGS} -o ${WINBINARY} -pkgdir /tmp/ -i

osx:  $(SOURCES)
	go get -v -d ../...
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=o64-clang CXX=o64-clang++ CGO_LDFLAGS="-w" go build ${LDFLAGS} -o ${OSXBINARY}

.PHONY: install
install:
	go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: list
list:
	cd .. && go list -f '{{ join .Imports "\n" }}'