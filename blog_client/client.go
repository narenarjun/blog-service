package main

import (
	"context"
	"fmt"
	"log"

	blogpbgen "github.com/narenarjun/blog-service/blogpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Blog Client Started ")

	opts := grpc.WithInsecure()
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer conn.Close()

	c := blogpbgen.NewBlogServiceClient(conn)


	// creating blog
	fmt.Println("Creating the blog")
	blog := &blogpbgen.Blog{
		AuthorId: "Naren",
		Title: "My first Blog",
		Content: "Content of the first blog post",
	}
	createBlogres, err := c.CreateBlog(context.Background(),&blogpbgen.CreateBlogRequest{
		Blog: blog,
	})
	if err != nil{
		log.Fatalf("Unexpected Error : %v\n", err)
		return
	}
	fmt.Printf("Blog has been created: %v\n",createBlogres)

}