FROM openshift/origin-release:golang-1.15 AS build
RUN mkdir -p /go-symlink-logfiles-watcher/cmd/go-symlink-logfiles-watcher
RUN mkdir -p /go-symlink-logfiles-watcher/pkg/symnotify
COPY ./cmd/go-symlink-logfiles-watcher/symlink-logfiles-watcher.go  /go-symlink-logfiles-watcher/cmd/go-symlink-logfiles-watcher/.
COPY ./pkg/symnotify/symnotify.go   /go-symlink-logfiles-watcher/pkg/symnotify/.
COPY ./go.mod  /go-symlink-logfiles-watcher/.
COPY ./go.sum  /go-symlink-logfiles-watcher/.
WORKDIR /go-symlink-logfiles-watcher
RUN go get -d -v ./...
RUN go mod vendor
EXPOSE 8080
RUN go build -o go-symlink-logfiles-watcher ./cmd/go-symlink-logfiles-watcher/symlink-logfiles-watcher.go
ENV pathoflogfiles=/var/log/containers/
ENV debugOn=true
ENV port=:2112
CMD echo "./go-symlink-logfiles-watcher -logfilespathname=${pathoflogfiles} -debug=${debugOn} -listeningport=${port}"
CMD ./go-symlink-logfiles-watcher -logfilespathname=${pathoflogfiles} -debug=${debugOn} -listeningport=${port}
