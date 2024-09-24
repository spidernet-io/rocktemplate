#!/usr/bin/make -f

all: run

.PHONY: run
run: build
	./main


.PHONY: build
build:
	rm bpf_*.go || true
	export GOPROXY="https://goproxy.io|https://goproxy.cn|direct" && \
		go mod tidy && \
		go generate
	# To further reduce runtime dependencies, add CGO_ENABLED=0 to go build and your application won't depend on libc
	CGO_ENABLED=0 go build


.PHONY: show_log
show_log:
	bpftool prog tracelog


.PHONY: show_map
show_map:
	bpftool map dump name map_floatip_v4
	bpftool map dump name map_backend_v4


.PHONY: show_prog
show_prog:
	bpftool cgroup tree /run/welan/cgroupv2
	bpftool prog show -f


.PHONY: tool
tool:
	apt-get update && apt-get install -y clang llvm gcc-multilib libbpf-dev linux-headers-$(uname -r) golang

