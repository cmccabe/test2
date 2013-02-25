#!/usr/bin/env bash

die() {
    echo $@
    exit 1
}

check_64bit() {
    foo=4611686018427387904
    [ $((foo+1)) != "4611686018427387905" ] && \
        die "your bash verison is too old to support 64-bit \
arithmetic.  Please upgrade"
}

[ $# -lt 1 ] && die "you must give an argument: the number of processes to spawn."
NUM_PROCS=$1
[ $NUM_PROCS -gt 0 ] || die "NUM_PROCS must be greater than 1 (it is $NUM_PROCS)"
check_64bit
DIR="`dirname $0`"
TD='TIME_DATA: user=%U, system=%S, elapsed=%e, CPU=%P, (%Xtext+%Ddata %Mmax)k, inputs=%I, outputs=%O, (%Fmajor+%Rminor)pagefaults, swaps=%W'

set -x

TDATA=858993459200
EDATA=$((TDATA / NUM_PROCS))
PROCS=`seq -w 1 $NUM_PROCS`

truncate -s $EDATA /tmp/f || die "truncate failed"
./bin/hadoop fs -rm '/*'
$DIR/dropCache || die "dropCache failed"
/usr/bin/time -f "${TD}" $DIR/raw-multi-copyFromLocal.sh $PROCS
rm -f /tmp/f
$DIR/dropCache || die "dropCache failed"
/usr/bin/time -f "${TD}" $DIR/raw-multi-hdfsCat.sh $PROCS
./bin/hadoop fs -rm '/*'

exit 0
