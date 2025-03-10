package main

import (
	"fmt"
	"log"

	"github.com/aceofkid/posix_mq"
	"github.com/google/flatbuffers/go"
	"github.com/johbau/rpc/go/Hello"
)

func main() {
	// Create or open the client message queue
	const clientFlags = posix_mq.O_RDONLY | posix_mq.O_CREAT
	clientMq, err := posix_mq.NewMessageQueue("/client_queue_hello", clientFlags, 0666, nil)
	if err != nil {
		log.Fatalf("Failed to create client message queue: %v", err)
	}
	defer clientMq.Close()
	defer clientMq.Unlink()

	// Create or open the server message queue
	const serverFlags = posix_mq.O_WRONLY
	serverMq, err := posix_mq.NewMessageQueue("/server_queue_hello", serverFlags, 0, nil)
	if err != nil {
		log.Fatalf("Failed to open server message queue: %v", err)
	}
	defer serverMq.Close()

	// Create a FlatBuffer request
	builder := flatbuffers.NewBuilder(1024)
	message := builder.CreateString("Hello from Go client!")
	Hello.HelloRequestStart(builder)
	Hello.HelloRequestAddRequestType(builder, Hello.RequestTypeGREETING)
	Hello.HelloRequestAddMessage(builder, message)
	request := Hello.HelloRequestEnd(builder)
	builder.Finish(request)

	// Send request to server
	err = serverMq.Send(builder.FinishedBytes(), 0)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Receive response
	msg, _, err := clientMq.Receive()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}

	// Parse the response
	response := Hello.GetRootAsHelloResponse(msg, 0)
	length := response.ResultLength()
	if length > 0 {
		fmt.Printf("Received response: %s\n", string(response.Result(0)))
	}
}
