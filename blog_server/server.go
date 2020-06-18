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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct{}

type blogitem struct{
	ID 			primitive.ObjectID 	`bson:"_id,omitempty"`
	AuthorID 	string 				`bson:"author_id"`
	Content 	string				`bson:"content"`
	Title 		string    			`bson:"title"`
}

func dataToBlog(data *blogitem) *blogpbgen.Blog{
	return &blogpbgen.Blog{
		Id: data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content: data.Content,
		Title: data.Title,
	}
}


func (*server) CreateBlog(ctx context.Context, req *blogpbgen.CreateBlogRequest) (*blogpbgen.CreateBlogResponse, error){
	
	fmt.Println("Create Blog request")
	blog := req.GetBlog()

	data := blogitem{
		AuthorID: blog.GetAuthorId(),
		Title: blog.GetTitle(),
		Content: blog.GetContent(),
	}

	res,err := collection.InsertOne(context.Background(),data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal Error : %v\n",err),
		)
	}

	oid,ok := res.InsertedID.(primitive.ObjectID)
	if !ok{
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID : %v\n",err),
		)
	}

	return &blogpbgen.CreateBlogResponse{
		Blog: &blogpbgen.Blog{
			Id: oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title: blog.GetTitle(),
			Content: blog.GetContent(),
		},
	} , nil
}

func (*server) ReadBlog( ctx context.Context, req *blogpbgen.ReadBlogRequest) (*blogpbgen.ReadBlogResponse, error){
	fmt.Println("Read Blog Request")

	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil{
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot Parse ID"),
		)
	}

	//  empty interface for the blog
	data := &blogitem{}
	filter := bson.M{"_id":oid}

	res := collection.FindOne(context.Background(),filter)
 	 if err := res.Decode(data); err != nil{
		  return nil, status.Errorf(
			  codes.NotFound,
			  fmt.Sprintf("Cannot find blog with the given ID : %v\n", err),
		  )
	  }

	  return &blogpbgen.ReadBlogResponse{
		  Blog: dataToBlog(data),
	  }, nil
}





func (*server) 	UpdateBlog(ctx context.Context, req *blogpbgen.UpdateBlogRequest) (*blogpbgen.UpdateBlogResponse, error) {
	fmt.Println("Update Blog Request")

	blog := req.GetBlog()
	oid , err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil{
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot Parse ID"),
		)
	}
	data := &blogitem{}
	filter := bson.M{"_id":oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil{
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with the given ID : %v\n", err),
		)
	}

	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, Updateerr := collection.ReplaceOne(context.Background(),filter,data)
	if err != nil{
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update Object in MongoDB: %v\n",Updateerr),
		)
	}

	return &blogpbgen.UpdateBlogResponse{
		Blog: dataToBlog(data),
	}, nil
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