#!/bin/bash

./bin/zot-linux-amd64 serve examples/config-test.json &
ZOT_PID=$!
# do other stuff

sleep 2
oras push 127.0.0.1:8080/hello/test:v1 statement.v2.json:application/vnd.uor.statement.v1+json --plain-http --verbose

#oras push 127.0.0.1:8080/hello/test:v2 nstatement.json:application/vnd.uor.statement.v1+json --plain-http --verbose

# When user ctrl+c, kill the zot process
wait $ZOT_PID
