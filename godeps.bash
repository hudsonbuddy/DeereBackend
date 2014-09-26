#!/bin/bash


export GOPATH=`pwd`

echo $GOPATH

go get -u github.com/gorilla/mux
go get -u github.com/gorilla/sessions
go get -u github.com/boj/redistore
go get -u labix.org/v2/mgo
