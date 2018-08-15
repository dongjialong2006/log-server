# The import path is where your repository can be found.
# To import subpackages, always prepend the full import path.
# If you change this, run `make clean`. Read more: https://git.io/vM7zV
IMPORT_PATH := log-server

V := 1 # When V is set, print commands and build progress.

# Space separated patterns of packages to skip in list, test, format.
IGNORED_PACKAGES := /vendor/


.PHONY: all
all: monitor

.PHONY: monitor
monitor: .GOPATH/.ok
	$Q go install -tags netgo $(IMPORT_PATH)

.PHONY: update
update: .GOPATH/.ok
	$Q glide up -v

.PHONY: docker
docker: 
	$Q docker build -f Dockerfile -t monitor:$(VERSION) .
	
##### =====> Utility targets <===== #####

.PHONY: clean test integration list cover format

clean:
	$Q rm -rf bin/log-server .GOPATH

test: .GOPATH/.ok
	$Q go test -p 1 $(if $V,-v) -i -race $(allpackages)
	
ifndef CI
	$Q go vet $(allpackages)
	$Q GODEBUG=cgocheck=2 go test -p 1 -race $(allpackages)
else
	$Q ( go vet $(allpackages); echo $$? ) | \
	    tee .GOPATH/test/vet.txt | sed '$$ d'; exit $$(tail -1 .GOPATH/test/vet.txt)
	$Q ( GODEBUG=cgocheck=2 go test -v -race $(allpackages); echo $$? ) | \
	    tee .GOPATH/test/output.txt | sed '$$ d'; exit $$(tail -1 .GOPATH/test/output.txt)
endif

list: .GOPATH/.ok
	@echo $(allpackages)

cover: .GOPATH/.ok
	@echo "NOTE: make cover does not exit 1 on failure, don't use it to check for tests success!"
	$Q rm -f .GOPATH/cover/*.out .GOPATH/cover/all.merged
	$(if $V,@echo "-- go test -coverpkg=./... -coverprofile=.GOPATH/cover/... ./...")
	@for MOD in $(allpackages); do \
		go test -coverpkg=`echo $(allpackages)|tr " " ","` \
			-coverprofile=.GOPATH/cover/unit-`echo $$MOD|tr "/" "_"`.out \
			$$MOD 2>&1 | grep -v "no packages being tested depend on"; \
	done
	$Q gocovmerge .GOPATH/cover/*.out > .GOPATH/cover/all.merged
ifndef CI
	$Q go tool cover -html .GOPATH/cover/all.merged
else
	$Q go tool cover -html .GOPATH/cover/all.merged -o .GOPATH/cover/all.html
endif
	@echo ""
	@echo "=====> Total test coverage: <====="
	@echo ""
	$Q go tool cover -func .GOPATH/cover/all.merged

format: .GOPATH/.ok
	$Q find .GOPATH/src/$(IMPORT_PATH)/ -iname \*.go | grep -v \
	    -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)) | xargs gofmt -w -s

BUILD            := $(shell git log --pretty=format:"%h" -1)
BRANCH 			 := $(shell git rev-parse --abbrev-ref HEAD)
VERSION          := $(shell git describe --abbrev=0 --tags)
DATE             := $(shell date -u '+%Y-%m-%d-%H%M UTC')
VERSION_FLAGS    := -ldflags='-X "$(IMPORT_PATH)/pkg/buildinfo.Version=$(VERSION)" -X "$(IMPORT_PATH)/pkg/buildinfo.BuildTime=$(DATE)" -X "$(IMPORT_PATH)/pkg/buildinfo.Branch=$(BRANCH)" -X "$(IMPORT_PATH)/pkg/buildinfo.Build=$(BUILD)"'

# cd into the GOPATH to workaround ./... not following symlinks
_allpackages = $(shell ( cd $(CURDIR)/.GOPATH/src/$(IMPORT_PATH) && \
    GOPATH=$(CURDIR)/.GOPATH/src/$(IMPORT_PATH) go list ./... 2>&1 1>&3 | \
    grep -v -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)) 1>&2 ) 3>&1 | \
    grep -v -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)))

# memoize allpackages, so that it's executed only once and only if used
allpackages = $(if $(__allpackages),,$(eval __allpackages := $$(_allpackages)))$(__allpackages)

export GOPATH := $(CURDIR)/.GOPATH

unexport GOBIN

Q := $(if $V,,@)

.GOPATH/.ok:
	$Q mkdir -p "$(dir .GOPATH/src)"
	$Q ln -s ../.. ".GOPATH/src"
	$Q mkdir -p bin
	$Q ln -s ../bin .GOPATH/bin
	$Q touch $@

.PHONY: bin/gocovmerge bin/goimports
bin/gocovmerge: .GOPATH/.ok
	@test -d ./vendor/github.com/wadey/gocovmerge || \
	    { echo "Vendored gocovmerge not found, try running 'make setup'..."; exit 1; }
	$Q go install $(IMPORT_PATH)/vendor/github.com/wadey/gocovmerge
bin/goimports: .GOPATH/.ok
	@test -d ./vendor/golang.org/x/tools/cmd/goimports || \
	    { echo "Vendored goimports not found, try running 'make setup'..."; exit 1; }
	$Q go install $(IMPORT_PATH)/vendor/golang.org/x/tools/cmd/goimports

# Based on https://github.com/cloudflare/hellogopher - v1.1 - MIT License
#
# Copyright (c) 2017 Cloudflare
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
