.PHONY: all clean build

# List of architectures and operating systems to support
ARCHS = amd64 arm64
OSES = linux darwin windows

# Mapping Go architectures to Zig architectures
ZIG_ARCHS_amd64 = x86_64
ZIG_ARCHS_arm64 = aarch64

# Mapping Go operating systems to Zig targets
ZIG_OSES_linux = linux-gnu
ZIG_OSES_darwin = macos
ZIG_OSES_windows = windows-gnu

# Default target
all: build

# Build libraries for all architecture and OS combinations
build:
	$(foreach os,$(OSES),\
        $(foreach arch,$(ARCHS),\
            CC="zig cc -target $(ZIG_ARCHS_$(arch))-$(ZIG_OSES_$(os))" \
            CGO_ENABLED=1 GOOS=$(os) GOARCH=$(arch) go build -o dnadesign/lib/libdnadesign_$(os)_$(arch).so -buildmode=c-shared lib.go;\
        )\
    )

clean:
	rm -f dnadesign/lib/libdnadesign_*.so

