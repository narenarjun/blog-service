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


 blogID :=	createBlog(c)

	readBlog(c, blogID)

}

func readBlog(c blogpbgen.BlogServiceClient ,bID string){

	fmt.Println("Reading the blog")

	// the following will show error because of the id is not in db
	_, err := c.ReadBlog(context.Background(),&blogpbgen.ReadBlogRequest{BlogId:"gdilkasjgakjgd" })
	if err != nil {
		fmt.Printf("Error while reading blog: %v\n", err)
	}

	readBlogreq := &blogpbgen.ReadBlogRequest{
		BlogId: bID,
	}

	res, err := c.ReadBlog(context.Background(),readBlogreq)

	if err != nil {
		fmt.Printf("Error while reading blog: %v\n", err)
	}

	fmt.Printf("Blog was read: %v\n", res)

}


func createBlog(c blogpbgen.BlogServiceClient ) string{
		// creating blog
		fmt.Println("Creating the blog")
		blog := &blogpbgen.Blog{
			AuthorId: "Arjun",
			Title: "My Infinite Blog",
			Content: "Content of the first blog post",
		}
		createBlogres, err := c.CreateBlog(context.Background(),&blogpbgen.CreateBlogRequest{
			Blog: blog,
		})
		if err != nil{
			log.Fatalf("Unexpected Error : %v\n", err)
			return ""
		}
		fmt.Printf("Blog has been created: %v\n",createBlogres)

		return createBlogres.GetBlog().GetId()
}

