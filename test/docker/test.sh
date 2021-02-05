#!/bin/bash

echo "Starting ovsdb-server..."
/usr/sbin/ovsdb-server --detach --remote=punix:/var/run/openvswitch/db.sock --pidfile=ovsdb-server.pid --remote=ptcp:6640 2> /dev/null

echo "Starting ovs-switchd..."
/usr/sbin/ovs-vswitchd --detach --verbose --pidfile 2> /dev/null

sleep 3

cd /CREBAS

go test ./...
