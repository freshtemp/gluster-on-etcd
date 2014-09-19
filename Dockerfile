FROM ubuntu:14.04

# All our build dependencies, in alphabetical order (to ease maintenance)
RUN apt-get update && apt-get install -y \
		software-properties-common \
		wget

RUN add-apt-repository -y ppa:semiosis/ubuntu-glusterfs-3.5

RUN apt-get update && apt-get install -y \
		glusterfs-server golang git pkg-config libzmq3-dev build-essential curl

ENV GOPATH /home 
RUN go get github.com/pebbe/zmq4
RUN go get github.com/coreos/go-etcd/etcd

EXPOSE 111 5555 24007 24009 24010 24011 24012 34865 34866 34867

ADD ./go/ /go/
