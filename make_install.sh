#!/usr/bin/env bash

die() {
    echo $@
    exit 1
}
[ $EUID == 0 ] || die "you must be root to run this."
set -e
curDir="`dirname $0`"
for f in "$curDir/setReadahead" "$curDir/dropCache"; do
    chown root:root $f
    chmod 4755 $f
done
