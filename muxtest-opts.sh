#!/usr/bin/env bash

[ $# -lt 2 ] && die "must supply two arguments: nthreads and total-megs" 
nthreads=$1
totalmegs=$2
shift
shift

set -x
set -e
DIR="`dirname $0`"
~/h/bin/hadoop fs -rm '/*' || true
$DIR/dropCache || die "dropCache failed"
echo "write test" 1>&2
echo "write test"
$DIR/muxtest.sh \
    -Dmuxtest.operation=write \
    -Dmuxtest.nthreads=$nthreads \
    -Dmuxtest.total.megs=$totalmegs \
    -Dmuxtest.hdfs.uri=hdfs://localhost:6000 \
    $@

$DIR/dropCache || die "dropCache failed"
echo "read test" 1>&2
echo "read test"
$DIR/muxtest.sh \
    -Dmuxtest.operation=read \
    -Dmuxtest.nthreads=$nthreads \
    -Dmuxtest.total.megs=$totalmegs \
    -Dmuxtest.hdfs.uri=hdfs://localhost:6000 \
    $@

~/h/bin/hadoop fs -rm '/*' || true
