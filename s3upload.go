package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
	"time"
)

const (
	mb = 1024 * 1024
)

var (
	awsProfile string
	up, down   bool
	part       int
)

func init() {
	flag.StringVar(&awsProfile, "p", "default", "aws profile used for the session")
	flag.IntVar(&part, "c", 10, "set concurency")
	flag.BoolVar(&up, "u", false, "upload file to s3 bucket")
	flag.BoolVar(&down, "d", false, "download object from s3 bucket")
	setupFlags(flag.CommandLine)
}

func setupFlags(f *flag.FlagSet) {
	f.Usage = func() {
		fmt.Printf(
			"\nExample: cat file | %s -u <bucket> <object_prefix> # read from stdin\n"+
				"Example: %s -d <bucket> <object_key> <file_path>\n\n"+
				"%s (-c) (-p) -d/-u <bucket> <key> (file_path)\n",
			os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           awsProfile,
	})
	if err != nil {
		awsError(err)
	}
	if up && down {
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	if len(flag.Args()) < 3 {
		if up {
			in := bufio.NewReaderSize(os.Stdin, part*mb)
			upload(sess, in)
		}
		if down {
			flag.CommandLine.Usage()
			os.Exit(1)
			//out := bufio.NewWriterSize(os.Stdout, part*mb)
			//download(sess, os.Stdout)
		}
	} else {
		if up {
			in, err := os.Open(flag.Arg(2))
			if err != nil {
				awsError(err)
			}
			upload(sess, in)
			in.Close()
		}
		if down {
			out, err := os.Create(flag.Arg(2))
			if err != nil {
				awsError(err)
			}
			download(sess, out)
			out.Close()
		}
	}
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

func upload(s *session.Session, in io.Reader) {

	file := s3manager.NewUploader(s)
	file.PartSize = int64(part * mb)
	file.Concurrency = part

	result, err := file.Upload(&s3manager.UploadInput{
		Bucket: aws.String(flag.Arg(0)),
		Key:    aws.String(keyName(flag.Arg(1))),
		Body:   in,
	})
	if err != nil {
		awsError(err)
	}
	fmt.Println("SUCCESS:", result.Location)
}

func download(s *session.Session, w io.WriterAt) {
	// https://gist.github.com/jboelter/ecfb08d6a18440ac16d93b5183aad207
	// is the solution how to do it
	// buff := &aws.WriteAtBuffer{}
	// and then we io.COPY to stout
	// need a way to be transparent

	file := s3manager.NewDownloader(s)
	file.PartSize = int64(part * mb)
	file.Concurrency = part

	_, err := file.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(flag.Arg(0)),
		Key:    aws.String(flag.Arg(1)),
	})
	if err != nil {
		awsError(err)
	}
}
