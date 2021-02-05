FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get -y update && apt-get -y upgrade
RUN apt-get -y upgrade
RUN apt-get -y install golang

CMD ["/CREBAS/test/docker/test.sh"]