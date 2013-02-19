#!/usr/bin/env bash

go build setReadahead.go
go build dropCache.go
go build test2.go subprocess.go
go build hdfsCat.go
