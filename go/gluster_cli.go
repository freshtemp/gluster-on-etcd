// Sends a command to a gluster go process at the given IP
// Usage: go run gluster_cli.go <ip_addr> <message>
// Scott Martin 9/2014

package main

import (
  zmq "github.com/pebbe/zmq4"
  "os"
  "fmt"
  "strings"
)

func main() {
  fmt.Println("Connecting to hello world server (" + os.Args[1] + ")...")
  requester, _ := zmq.NewSocket(zmq.REQ)
  defer requester.Close()
  requester.Connect("tcp://"+os.Args[1]+":5555")
  msg := strings.Join(os.Args[2:], " ")
  fmt.Println("Sending ", msg)
  requester.Send(msg, 0)
  reply, _ := requester.Recv(0)
  fmt.Println("Response:\n", reply)
}
