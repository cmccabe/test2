#!/usr/bin/env bash

die() {
    echo $@
    exit 1
}
mkdir -p /home/cmccabe/h/share/hadoop/common/ /home/cmccabe/h/share/hadoop/hdfs/
source /home/cmccabe/cmccabe-hbin/jarjar.sh /home/cmccabe/h/ \
    /home/cmccabe/h/share/hadoop/common/ \
    /home/cmccabe/h/share/hadoop/hdfs/
export CLASSPATH="$CLASSPATH:./build/jar/HioBench.jar"
export LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/home/cmccabe/h/lib/native"
pushd ~/src/hio_test
ant clean compile jar || die "ant failed"
sleep 30 # sleep to avoid coming up before the cluster is ready.  TODO: put something better here
TD='TIME_DATA: user=%U, system=%S, elapsed=%e, CPU=%P, (%Xtext+%Ddata %Mmax)k, inputs=%I, outputs=%O, (%Fmajor+%Rminor)pagefaults, swaps=%W'
/usr/bin/time -f "$TD" java "$@" com.cloudera.HioBench || die "java failed"
popd
