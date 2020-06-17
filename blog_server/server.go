package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	blogpbgen "github.com/narenarjun/blog-service/blogpb"
	"google.golang.org/grpc"
)

type server struct{}

func main(){
	// if our program crashes, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)


	fmt.Println("Blog Service started")

	// ! 50051 is the default port for grpc
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failes to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpbgen.RegisterBlogServiceServer(s , &server{})

	go func(){
		fmt.Println("starting Server...")

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve : %v", err)
		}
	}()

	//  wait for ctrl + c to exit
	ch := make(chan os.Signal,1)
	signal.Notify(ch, os.Interrupt)

	// block untill signal is received
	<-ch
	fmt.Println("Stopping the server...")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("END of Program")
}