package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	blogpbgen "github.com/narenarjun/blog-service/blogpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

var collection *mongo.Collection

type server struct{}

type blogitem struct{
	ID 			primitive.ObjectID 	`bson:"_id,omitempty"`
	AuthorID 	string 				`bson:"author_id"`
	Content 	string				`bson:"content"`
	Title 		string    			`bson:"title"`
}

func main(){
	// if our program crashes, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB")


	// ! the mongodb  connection url must be supplied for it to work properly

	// * connection to MongoDB
	client, err :=  mongo.NewClient(options.Client().ApplyURI("mongodb+srv://<username>:<password>@grpc1cluster-rqxlq.mongodb.net/test"))

	if err != nil { 
		log.Fatalf("Error while connecting to mongodb: %v\n",err)
		return 
	}
		
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil { 
		log.Fatalf("Error while acquiring connection to database: %v\n",err)
		return 
	}

	collection = client.Database("mydb").Collection("blog")

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
	fmt.Println("Closing MongoDb connection")
	client.Disconnect(ctx)
	fmt.Println("END of Program")
}