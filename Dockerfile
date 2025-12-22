FROM golang:1.25 as builder

LABEL maintainer="Bruce Xiu <bruce.xiu@ballet.com>"

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential git ca-certificates pkg-config \
    libzmq3-dev libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev \
    liblz4-dev libzstd-dev clang lld llvm && \
    rm -rf /var/lib/apt/lists/*

ENV ROCKSDB_VERSION=v9.10.0

#RUN mkdir /build
# install rocksdb https://github.com/facebook/rocksdb/blob/master/INSTALL.md
RUN git clone -b $ROCKSDB_VERSION --depth 1 https://github.com/facebook/rocksdb.git /opt/rocksdb && \
    cd /opt/rocksdb && \
    CFLAGS=-fPIC CXXFLAGS=-fPIC LDFLAGS="-latomic" CC=clang CXX=clang++ make -j"$(nproc)" release

RUN mkdir -p /build/lib && \
    cp -av /opt/rocksdb/librocksdb.so* /build/lib/ && \
    strip /opt/rocksdb/ldb /opt/rocksdb/sst_dump && \
    cp /opt/rocksdb/ldb /opt/rocksdb/sst_dump /build/ && \
    ldd /build/lib/librocksdb.so* || true


WORKDIR /build

COPY . ./
RUN go mod download


ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-I/opt/rocksdb/include"
ENV CGO_LDFLAGS="-L/opt/rocksdb -lrocksdb -lstdc++ -lm -lz -ldl -lbz2 -lsnappy -llz4 -lzstd -latomic"
RUN go build -mod=readonly -tags rocksdb_9_10 -trimpath -v -o /build/blockbook

####### Start a new stage from debian #######
FROM debian:trixie-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates libatomic1 \
    libstdc++6 zlib1g libbz2-1.0 libsnappy1v5 liblz4-1 libzstd1 libgflags2.2 \
    libzmq5 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/lib/librocksdb.so* /usr/local/lib/
RUN ldconfig

WORKDIR /blockbook

COPY --from=builder /build/blockbook /blockbook/bin/blockbook
COPY --from=builder /build/static /blockbook/static/

EXPOSE 9034
EXPOSE 9134
EXPOSE 9090

ENTRYPOINT ["/blockbook/bin/blockbook", "-workers=1", "-dbcache=300", "-sync", "-blockchaincfg=/blockbook/blockchaincfg.json", "-datadir=/blockbook/data", "-internal=:9034", "-public=:9134", "-explorer=", "-logtostderr", "-prof=:9090", "-dbstatsperiod=0", "-resyncindexperiod=61091", "-resyncmempoolperiod=2999"]
