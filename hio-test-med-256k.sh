#!/usr/bin/env bash

set -x
set -e
DIR="`dirname $0`"
$DIR/hio-test.sh \
    -Dhio.nthreads=10 -Dhio.ngigs.to.read=2 \
    -Dhio.read.chunk.bytes=262144 -Dhio.ngigs.in.file=800 \
    -Dhio.hdfs.uri=hdfs://localhost:6000 -Dhio.hdfs.file.name=/hio
