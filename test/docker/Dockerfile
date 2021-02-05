FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get -y update && apt-get -y upgrade
RUN apt-get -y upgrade
RUN apt-get -y install curl
RUN apt-get -y install golang
RUN apt-get -y install openvswitch-switch openvswitch-common
RUN apt-get -y install iproute2

# special thanks to https://github.com/globocom/docker-openvswitch
# Create database and pid file directory
RUN /usr/bin/ovsdb-tool create /etc/openvswitch/conf.db
RUN mkdir -pv /var/run/openvswitch/

CMD ["/CREBAS/test/docker/test.sh"]
