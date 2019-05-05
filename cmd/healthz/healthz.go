package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"encoding/binary"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/onlinecity/ocmg-go-health/pkg/oc_pb_rpc"
	zmq "github.com/pebbe/zmq4"
)

// parseReply parses the healthz reply according to OC RPC protocol
// doing this without a full rpc implementation is a bit involved, but keep deps simple
func parseReply(reply [][]byte) {
	if len(reply) < 1 { // ZMQ error
		log.Panic("malformed reply from server")
	}
	if len(reply[0]) != 4 { // OC RPC violation
		log.Panic("malformed reply from server")
	}
	var parts int32
	if err := binary.Read(bytes.NewReader(reply[0]), binary.LittleEndian, &parts); err != nil {
		log.Fatalln(err.Error())
	}
	if parts == 0 { // void reply, in this context OK
		log.Println("healthz: OK")
		os.Exit(0)
	}
	if parts > 0 { // our contract states void or exception, no rpc body
		log.Fatalln("unexpected reply from server")
	}
	if parts == -1 { // an exception was raised
		if len(reply) != 2 { // OC RPC violation
			log.Panic("malformed reply from server")
		}
		// try to parse the exception
		ex := &oc_pb_rpc.Exception{}
		err := proto.Unmarshal(reply[1], ex)
		if err != nil {
			log.Fatalf("unexpected error reply from server %q", reply[1])
		}
		var u string
		if uu, err := uuid.FromBytes(ex.IncidentUuid); err == nil {
			u = uu.String()
		} else {
			u = fmt.Sprintf("%x", ex.IncidentUuid)
		}
		log.Fatalf("got exception(%04X): %q -- uuid:%v - vars:%v",
			ex.Code, *ex.Message, u, ex.Variables)
	}
	log.Panic("malformed reply from server")
}

func main() {
	var endpoint = flag.String("endpoint", "tcp://localhost:7200", "where to connect")
	var retries = flag.Uint("retries", 3, "how many retries")
	var timeout = flag.Uint("timeout", 2000, "timeout in ms")
	flag.Parse()

	log.Printf("connecting to %q\n", *endpoint)
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		panic(err)
	}
	if err := client.Connect(*endpoint); err != nil {
		log.Fatalln(err.Error())
	}

	poller := zmq.NewPoller()
	poller.Add(client, zmq.POLLIN)

	poll_timeout := time.Duration(*timeout) * time.Millisecond
	retries_left := *retries
	for retries_left > 0 {
		//  We send a request, then we work to get a reply
		if _, err := client.SendMessage("healthz"); err != nil {
			log.Fatalln(err.Error())
		}

		//  Poll socket for a reply, with timeout
		sockets, err := poller.Poll(poll_timeout)
		if err != nil {
			log.Fatal(err)
		}

		//  We got a reply from the server
		if len(sockets) > 0 {
			reply, err := client.RecvMessageBytes(0)
			if err != nil {
				log.Println(err.Error())
			}
			client.Close()
			parseReply(reply)
			return
		}

		// retry logic
		retries_left--
		if retries_left == 0 {
			log.Fatalln("server seems to be offline, abandoning")
		} else {
			log.Println("no response from server, retrying...")
			//  Old socket is confused; close it and open a new one
			client.Close()
			client, _ = zmq.NewSocket(zmq.REQ)
			if err := client.Connect(*endpoint); err != nil {
				log.Fatalln(err.Error())
			}
			// Recreate poller for new client
			poller = zmq.NewPoller()
			poller.Add(client, zmq.POLLIN)
		}
	}
}
