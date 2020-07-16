package main

// 20200715
// read arbitrary data from linux pipe and create new bkp file in s3
import (
	"bufio"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"time"
)

const (
	mb   = 1024 * 1024
	part = 10
)

func init() {
	if len(os.Args) != 3 {
		help()
	}
}

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		awsError(err)
	}

	file := s3manager.NewUploader(sess)
	file.PartSize = part * mb
	file.Concurrency = part

	if isInputFromPipe() {
		result, err := file.Upload(&s3manager.UploadInput{
			Bucket: aws.String(os.Args[1]),
			Key:    aws.String(keyName(os.Args[2])),
			Body:   bufio.NewReaderSize(os.Stdin, part*mb),
		})
		if err != nil {
			awsError(err)
		}
		fmt.Println("SUCCESS:", result.Location)
	} else {
		help()
	}
}

func help() {
	fmt.Fprintf(os.Stderr, "Example: cat file | %s <bucket> <object>\n",
		os.Args[0])
	os.Exit(1)
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func awsError(err error) {
	if awsErr, ok := err.(awserr.Error); ok {
		fmt.Fprintf(os.Stderr, "Error: %s %s\n", awsErr.Code(), awsErr.Message())
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			fmt.Fprintf(os.Stderr, "ReqErr: %d %s\n", reqErr.StatusCode(),
				reqErr.RequestID())
		}
	} else {
		fmt.Println(err.Error())
	}
	os.Exit(1)
}

func keyName(fileName string) string {
	t := time.Now()
	return t.Format("2006/01/02/") + fileName + "_" + t.Format("15_04_05")
}
