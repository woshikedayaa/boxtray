NAME = boxtray
WORKDIR = .
VERSION = $(shell cat $(WORKDIR)/.VERSION)-commit-$(shell git rev-parse HEAD)
OUTPUT = $(WORKDIR)/bin/$(NAME)
VERSION_FLAG = -X github.com/woshikedayaa/${NAME}/cmd/${NAME}/metadata.Version=$(VERSION)
LDFLAGS = $(VERSION_FLAG) -s -w

# Detect operating system
ifeq ($(GOOS),windows)
OUTPUT := $(OUTPUT).exe
endif

# Check if Clang compiler is available
CLANG_CHECK := $(shell which clang 2>/dev/null)
ifdef CLANG_CHECK
	CC = clang
	CXX = clang++
endif

# Qt-related CGO settings
CGO_ENABLED = 1

# Enable C++17 support for qt6
CGO_CXXFLAGS += -std=c++17

# Add Qt library paths (may need adjustment based on actual installation)
ifdef QTDIR
	CGO_CXXFLAGS += -I$(QTDIR)/include
	CGO_LDFLAGS += -L$(QTDIR)/lib
endif

# Main compilation parameters
PARAMS = -trimpath -ldflags "$(LDFLAGS)" -v -o $(OUTPUT)

.PHONY: build
build: clean
	CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_CXXFLAGS=$(CGO_CXXFLAGS) CGO_ENABLED=$(CGO_ENABLED) \
go build $(PARAMS) $(WORKDIR)

LINUX_TARGETS = amd64 386 arm arm64
.PHONY: $(LINUX_TARGETS)
$(LINUX_TARGETS):
	@mkdir -p $(WORKDIR)/bin/release/linux_$@
	GOOS=linux GOARCH=$@ CC=$(CC) \
CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_CXXFLAGS=$(CGO_CXXFLAGS) CGO_ENABLED=$(CGO_ENABLED) \
go build $(PARAMS) -o $(WORKDIR)/bin/release/linux_$@/$(NAME) $(WORKDIR)

.PHONY: windows
windows:
	@mkdir -p $(WORKDIR)/bin/release/windows_amd64
	GOOS=$@ GOARCH=amd64 CC=$(CC) \
CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_CXXFLAGS=$(CGO_CXXFLAGS) CGO_ENABLED=$(CGO_ENABLED) \
go build $(PARAMS) -o $(WORKDIR)/bin/release/linux_$@/$(NAME) $(WORKDIR)

.PHONY: release
release: clean $(LINUX_TARGETS) windows

.PHONY: clean
clean:
	@rm -rf $(OUTPUT)
	@rm -rf $(WORKDIR)/bin/release