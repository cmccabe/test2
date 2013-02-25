#!/usr/bin/env bash

for i in $@; do
    /usr/bin/time -f "${TD}" ./bin/hadoop fs -copyFromLocal /tmp/f /tmp.$i &
done
wait
