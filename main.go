package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/radovskyb/watcher"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	log.Println("Starting file Monitor")
	path := os.Getenv("FOLDER")
	log.Printf("Monitoring Folder: %v", path)
	log.Println("Checking folder for any files")
	dir, err := os.Open(path)
	if err != nil {
		log.Fatalf("Unable to open folder: %v", err)
	}

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Fatalf("Unable to list files: %v", err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name(),"."){
			continue
		}

		sendFile(watcher.Event{
			Path:     path+"/"+f.Name(),
			FileInfo: f,
		})

	}

	w := watcher.New()
	w.FilterOps(watcher.Create, )
	err = w.Add(os.Getenv("FOLDER"))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				sendFile(event)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.Start(time.Second * 60); err != nil {
		log.Fatalln(err)
	}

}

func sendFile(event watcher.Event) {
	if strings.HasPrefix(event.Path,"."){
		return
	}
	log.Printf("Sending File: %v", event.Path)
	client := http.DefaultClient
	transport := http.DefaultTransport
	transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client.Transport = transport
	config := aws.Config{LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody), HTTPClient: client}
	sess, _ := session.NewSession(&config)
	f, err := os.Open(event.Path)
	if err != nil {
		log.Fatalf("failed to open file %q, %v", event.Path, err)
	}
	var size  = event.Size()

	buffer := make([]byte, size)
	_, err = f.Read(buffer)
	if err != nil {
		log.Fatalf("failed to read file %q, %v", event.Path, err)
	}

	// Upload the file to S3.
	_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(os.Getenv("BUCKET")),
		Key:                  aws.String(event.Name()),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
	})
	if err != nil {
		log.Fatalf("failed to upload file, %v", err)
	}
	err = os.Remove(event.Path)
	if err != nil {
		log.Fatalf("failed to upload file, %v", err)
	}
	fmt.Printf("file uploaded")

}
