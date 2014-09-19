#!/bin/bash
# This script handles startup of gluster FS containers inside CoreOS
# Scott Martin 9/2014

# TODO: This code is out of date with run_server.sh and run_client.sh...
# it should probably execute these other scripts to issue commands.


# TODO: actual IPs for other servers
CORE1=104.130.3.50
CORE2=111.111.111.111 
CORE3=222.222.222.222

ETCD_ENDPOINT=http://$(ifconfig docker0 | grep 'inet ' | awk '{ print $2}'):4001
echo etcd endpoint: $ETCD_ENDPOINT

GLUSTER_CONTAINER="gluster_daemon"
GLUSTER_IMAGE="gluster"
GLUSTER_PORT=5555
container_id=$(docker inspect --format '{{ .Config.Hostname }}' $GLUSTER_CONTAINER)

CLIENT_CMD="docker run -t -i $GLUSTER_IMAGE go run /home/gluster_cli.go"

run_gluster_cmd() {
  # Use inspect to get server container's IP
  srv_ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' $GLUSTER_CONTAINER)
  echo Using $srv_ip for gluster daemon IP

  # Start up client process to connect to server
  result=$($CLIENT_CMD $srv_ip "$@" $srv_ip)
  echo result
}

cmd="$@"
CMD_START="start"
CMD_JOIN="join"
CMD_NEWVOL="volume create"

if [ "$cmd" == "$CMD_START" ]; then

  if [ "$container_id" = "" ]; then
    echo "Starting"
    docker run -i -t --name $GLUSTER_CONTAINER $GLUSTER_IMAGE sh -c 'glusterd && go run /home/gluster_srv.go $ETCD_ENDPOINT'
  else
    echo "Already started"
    exit 1
  fi
elif [ "$cmd" == "$CMD_JOIN" ]; then
  if [ "$container_id" = "" ]; then
    echo "Gluster not running" 
    exit 1
  else 
    echo "Attempting to join gluster pool"
    # TODO: Lookup etcd IP
    run_gluster_cmd $CMD_JOIN 
  fi
elif [[ $cmd == $CMD_NEWVOL* ]]; then
  echo "New volume"
else
  echo "Usage:"
  echo "$0 $CMD_START" 
  echo "$0 $CMD_JOIN"
  echo "$0 $CMD_NEWVOL <options>"
  exit 1
fi

echo $cmd

