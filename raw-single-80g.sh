#!/usr/bin/env bash

die() {
    echo $@
    exit 1
}

DIR="`dirname $0`"
TD='TIME_DATA: user=%U, system=%S, elapsed=%e, CPU=%P, (%Xtext+%Ddata %Mmax)k, inputs=%I, outputs=%O, (%Fmajor+%Rminor)pagefaults, swaps=%W'

set -x

truncate -s 85899345920 /tmp/f || die "truncate failed"
$DIR/dropCache || die "dropCache failed"
/usr/bin/time -f "${TD}" ./bin/hadoop fs -copyFromLocal /tmp/f /f
$DIR/dropCache || die "dropCache failed"
/usr/bin/time -f "${TD}" ./bin/hadoop fs -cat /f > /dev/null

./bin/hadoop fs -rm /f
rm -f /tmp/f
