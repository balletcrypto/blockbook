FROM golang:1.17.1 as builder

LABEL MAINTAINER="Andy Wang <andy.wang@mintery.com>"

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y build-essential git wget pkg-config lxc-dev libzmq3-dev \
                       libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev \
                       liblz4-dev graphviz && \
    apt-get clean

ENV ROCKSDB_VERSION=v6.22.1
ENV CGO_CFLAGS="-I/opt/rocksdb/include"
ENV CGO_LDFLAGS="-L/opt/rocksdb -lrocksdb -lstdc++ -lm -lz -ldl -lbz2 -lsnappy -llz4"

RUN mkdir /build
# install rocksdb https://github.com/facebook/rocksdb/blob/master/INSTALL.md
RUN cd /opt && git clone -b $ROCKSDB_VERSION --depth 1 https://github.com/facebook/rocksdb.git && \
    cd /opt/rocksdb && \
    CFLAGS=-fPIC CXXFLAGS=-fPIC make -j 4 release

RUN strip /opt/rocksdb/ldb /opt/rocksdb/sst_dump && \
    cp /opt/rocksdb/ldb /opt/rocksdb/sst_dump /build


WORKDIR /build

COPY . ./
RUN go mod download


RUN go build -mod=readonly -tags rocksdb_6_16 -v -o blockbook

####### Start a new stage from debian #######
#FROM debian:buster-slim
FROM debian:bullseye

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y build-essential git wget pkg-config lxc-dev libzmq3-dev \
                       libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev \
                       liblz4-dev graphviz && \
    apt-get clean

WORKDIR /blockbook

COPY --from=builder /build/blockbook /blockbook/bin/
COPY --from=builder /build/static /blockbook/static/

EXPOSE 9034
EXPOSE 9134
EXPOSE 9090

ENTRYPOINT ["/blockbook/bin/blockbook", "-workers=1", "-dbcache=300", "-sync", "-blockchaincfg=/blockbook/blockchaincfg.json", "-datadir=/blockbook/data", "-internal=:9034", "-public=:9134", "-explorer=", "-logtostderr", "-prof=:9090", "-dbstatsperiod=0", "-resyncindexperiod=61091"]
