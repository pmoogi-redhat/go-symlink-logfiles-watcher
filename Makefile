all: build

copy_files: 
	cp -r ./cmd Docker/.
	cp -r ./pkg Docker/.
	cp  ./go.*  Docker/.

build:  copy_files
	hack/build-component-image.sh Docker  docker.io/cognetive/go-symlink-logfiles-watcher_v0
