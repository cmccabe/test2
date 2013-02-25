#!/usr/bin/env bash

set -x
dir="`dirname $0`"
"$dir/hio-test.sh" \
    -Dhio.nthreads=100 -Dhio.nmegs.to.read=512 \
    -Dhio.read.chunk.bytes=131072 -Dhio.ngigs.in.file=800 \
    -Dhio.hdfs.uri=hdfs://localhost:6000 -Dhio.hdfs.file.name=/hio \
    -Dhio.hdfs.test.type=sequential \
    -Xmx3g -Xss256k

 #800 \
