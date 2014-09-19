ip_addr=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' $1)

if [ "$ip_addr" == "<no value>" ] || [ "$ip_addr" == "" ] ; then
  echo "Couldn't find IP for $1"
  exit 1
fi
echo Container IP: $ip_addr

# Remove initial container id
shift 1

docker run -t -i gluster go run /go/gluster_cli.go $ip_addr $@
