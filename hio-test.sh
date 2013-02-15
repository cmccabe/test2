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
java "$@" com.cloudera.HioBench || die "java failed"
popd
