// Gluster daemon server
// Handles volume creation and peer probing into the main pool
// Scott Martin 9/2014

package main

import (
  zmq "github.com/pebbe/zmq4"
  etcd "github.com/coreos/go-etcd/etcd"
  "fmt"
  "log"
  "bytes"
  "os"
  "os/exec"
  "os/signal"
  "strings"
  "net"
)

const POOL_KEY string = "/glusterpool"

var etcd_client *etcd.Client

func pool_register(hostip string) {
  // Enters the hostip entry into etcd at POOL_KEY

  _, err := etcd_client.Set(POOL_KEY+"/"+hostip, hostip, 0)
  if err != nil {
    log.Fatal(err)
  }
  log.Printf("Registered:" + hostip)
}

func pool_deregister(hostip string) {
  // Removes the hostip entry from etcd

  if r := recover(); r != nil {
    fmt.Println("Recovered in deregister", r)
  }
  _, err := etcd_client.Delete(POOL_KEY+"/"+hostip, false)
  if err != nil {
    log.Printf("Deregister from etcd failed... key likely was never inserted")
    log.Printf(err.Error()) 
  } else {
    log.Printf("Unregistered:" + hostip)
  }
}

func execute(name string, arg ...string) (string, error) {
  // Wraps exec.Command, returning output and error codes and
  // writing stderr to console

  cmd := exec.Command(name, arg...)
  var out bytes.Buffer
  var serr bytes.Buffer
  cmd.Stdout = &out
  cmd.Stderr = &serr
  err := cmd.Run()
  log.Printf(serr.String())
  return out.String(), err
}

func peer_probe(ip string) (bool) {
  // Attempt to join an outsider gluster instance to this peer's pool.
  // Also inserts this peer into etcd
  log.Printf("Executing gluster peer probe "+ip)
  out, err := execute("gluster", "peer", "probe", ip)
  if err != nil {
    log.Printf("Probe failed!")
    log.Printf(err.Error())
    return false
  } else {
    log.Printf(out)  
    pool_register(ip)
    return true
  }
}

func volume_create(fields []string) {
  // TODO: Allow auto-population of volume's servers (and quantity)
  out, err := execute("gluster", "volume " + strings.Join(fields[1:], " "))
  if err != nil {
    log.Printf("Volume create failed")
    fmt.Printf(err.Error())
  }
  log.Println(out)
}

func get_eth0_ip() (string) {
  // Looks for the IP of eth0, used for self-registration
  // This will likely be irrelevant once we change to a 
  // pull-style peer probe request through etcd.

  interfaces, _ := net.Interfaces()
  for _, inter := range interfaces {
    if inter.Name != "eth0" {
      continue
    }
    if addrs, err := inter.Addrs(); err == nil {
      return strings.Split(addrs[0].String(), "/")[0]
    } 
  }
  log.Fatal("Couldn't resolve host IP")
  return ""
}



func main() {
  fmt.Println(strings.Join(os.Args, " "))
  etcd_server_ip := os.Args[1]
  etcd_client = etcd.NewClient([]string{etcd_server_ip})

  // Hostname is needed for insertion/removal from etcd
  hostip := get_eth0_ip()
  log.Println(hostip)
  log.Printf("Host IP: "+hostip)
  defer pool_deregister(hostip)

  // This handles Ctrl+C actions, lets us deregister cleanly
  // TODO: Remove this once converted to pull-style peer probe?
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  go func() {
    for sig := range c {
      fmt.Println("Got signal", sig)
      pool_deregister(hostip)
      os.Exit(1)
    } 
  }()

  responder, _ := zmq.NewSocket(zmq.REP)
  defer responder.Close()
  responder.Bind("tcp://*:5555")

  for {
    //  Wait for next request from client
    msg, _ := responder.Recv(0)
    fmt.Println("Got message \""+msg+"\"")
    fields := strings.Fields(msg)

    switch fields[0] {
      case "join":
        if len(fields) != 2 {
          log.Println("Invalid number of arguments for join")
          responder.Send("Invalid command", 0)
          continue
        }

        if fields[1] == hostip {
          log.Println("Joining self")
          pool_register(hostip)
          responder.Send("Made new pool", 0)
        } else {
          log.Println("Attepting pool join (target "+fields[1]+")")
          if peer_probe(fields[1]) {
            log.Println("Joined "+fields[1])
            responder.Send("Joined "+fields[1], 0)
          } else {
            log.Println("Couldn't join "+fields[1])
            responder.Send("Couldn't join "+fields[1], 0)
          }
        }
      case "volume":
        fmt.Println("Attempting volume create")
        volume_create(fields)
        responder.Send("Created volume", 0)
      default:
        log.Println("Unknown command \""+msg+"\"")
        responder.Send("Unknown command \""+msg+"\"", 0)
    }
  }
}
