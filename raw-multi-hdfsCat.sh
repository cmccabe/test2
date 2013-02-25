#!/usr/bin/env bash

for i in $@; do
    /usr/bin/time -f "${TD}" ./bin/hadoop fs -cat /tmp.$i > /dev/null &
done
wait
