#! /bin/bash

cd agent
go generate
if (( $? )); then
    exit 1
fi
cd ..
go build