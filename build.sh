#! /bin/bash

cd mcservice
go generate
if (( $? )); then
    exit 1
fi
cd ..
go build