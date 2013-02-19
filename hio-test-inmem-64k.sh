#!/usr/bin/env bash

set -x
set -e
DIR="`dirname $0`"
$DIR/hio-test.sh \
    -Dhio.nthreads=10 -Dhio.ngigs.to.read=1 \
    -Dhio.read.chunk.bytes=65536 -Dhio.ngigs.in.file=40 \
    -Dhio.hdfs.uri=hdfs://localhost:6000 -Dhio.hdfs.file.name=/hio
