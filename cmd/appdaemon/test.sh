#!/bin/bash

TEST_DIR=$(cd $(dirname $0); pwd)

go build

echo "Test noSIGTERM"
./appdaemon testMode noSIGTERM &
PID=$!
sleep 1
kill -TERM $PID 2> /dev/null
sleep 1
EXIT_CODE=`cat /tmp/appdaemon_test`
if [ $EXIT_CODE != "0" ];then
	echo "Test Failed"
	exit 1
fi

echo "Test SIGTERM"
./appdaemon testMode SIGTERM &
PID=$!
sleep 1
kill -TERM $PID
sleep 1
EXIT_CODE=`cat /tmp/appdaemon_test`
if [ $EXIT_CODE == "0" ];then
	echo "Test Failed"
	exit 1
fi
