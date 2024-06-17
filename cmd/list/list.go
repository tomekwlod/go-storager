package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	gst "github.com/tomekwlod/go-storager/google_storage"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("this command expects exactly 2 arguments; something like:\n\n\t%s\n\n", "./list /env/file/.env remote/path")
	}

	env := os.Args[1]

	err := godotenv.Load(env)
	if err != nil {
		log.Fatalf("file '%s' cannot be found\n", env)
	}

	remotePath := os.Args[2]

	googleStorage := gst.Setup(context.Background(), os.Getenv("GCS_SA_KEY_JSON"), os.Getenv("GCS_BUCKET"))

	l, err := googleStorage.List(remotePath)
	if err != nil {
		panic(err)
	}

	for _, v := range l {
		fmt.Printf("> %s\n", v.Path)
	}
	fmt.Println("> all files:", len(l))
}
