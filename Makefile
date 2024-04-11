init:
	git submodule update --init --depth 1 ollamax
	make -C ollamax

build: init
	go build .