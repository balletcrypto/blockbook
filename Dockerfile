FROM golang:1.15.7 as builder

LABEL MAINTAINER="Andy Wang <andy.wang@mintery.com>"

RUN apt-get update && \
   apt-get upgrade -y && \
   apt-get install -y git wget && \
   apt-get install -y libzmq3-dev libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev

ENV ROCKSDB_VERSION=v6.13.3
ENV CGO_CFLAGS="-I/opt/rocksdb/include"
ENV CGO_LDFLAGS="-L/opt/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4"

# install rocksdb https://github.com/facebook/rocksdb/blob/master/INSTALL.md
RUN cd /opt && git clone -b $ROCKSDB_VERSION --depth 1 https://github.com/facebook/rocksdb.git && \
    cd /opt/rocksdb && \
    CFLAGS=-fPIC CXXFLAGS=-fPIC make -j 2 static_lib

WORKDIR /build

COPY go.* ./

RUN go mod download

COPY . ./

RUN go build -mod=readonly -v -o blockbook

####### Start a new stage from debian #######
FROM debian:buster-slim

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzmq3-dev

WORKDIR /blockbook

COPY --from=builder /build/blockbook /blockbook/bin/
COPY --from=builder /build/static /blockbook/static/
COPY --from=builder /build/build/text/ /build/build/text/

EXPOSE 9034
EXPOSE 9134
EXPOSE 9090

ENTRYPOINT ["/blockbook/bin/blockbook", "-workers=1", "-dbcache=300", "-sync", "-blockchaincfg=/blockbook/blockchaincfg.json", "-datadir=/blockbook/data", "-internal=:9034", "-public=:9134", "-explorer=", "-logtostderr", "-prof=:9090", "-dbstatsperiod=0", "-resyncindexperiod=61091"]
