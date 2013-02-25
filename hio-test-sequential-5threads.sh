#!/usr/bin/env bash

set -x
dir="`dirname $0`"
"$dir/hio-test.sh" \
    -Dhio.nthreads=5 -Dhio.ngigs.to.read=3 \
    -Dhio.read.chunk.bytes=1048576 -Dhio.ngigs.in.file=800 \
    -Dhio.hdfs.uri=hdfs://localhost:6000 -Dhio.hdfs.file.name=/hio \
    -Dhio.hdfs.test.type=sequential
