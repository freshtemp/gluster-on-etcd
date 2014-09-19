
etcd_host=http://$(ifconfig eth1 | grep 'inet ' | awk '{ print $2}'):4001
echo "Running gluster server (etcd server @ $etcd_host)"

docker run -t -i gluster sh -c 'glusterd && go run /go/gluster_srv.go '"$etcd_host"
