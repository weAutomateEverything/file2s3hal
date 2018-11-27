package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/prometheus/common/log"
	"github.com/radovskyb/watcher"
	"os"
	"time"
)

func main(){
	w := watcher.New()
	w.FilterOps(watcher.Create)
	err := w.Add(os.Getenv("FOLDER"))
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
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	f, err  := os.Open(event.Path+event.Name())
	if err != nil {
		log.Fatalf("failed to open file %q, %v", event.Name(), err)
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("BUCKET")),
		Key:    aws.String(event.Name()),
		Body:   f,
	})
	if err != nil {
		log.Fatalf("failed to upload file, %v", err)
	}
	os.Remove(event.Name())
	fmt.Printf("file uploaded to, %s\n", result.Location)

}



