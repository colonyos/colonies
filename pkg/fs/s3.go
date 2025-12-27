package fs

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type S3Client struct {
	mc         *minio.Client
	BucketName string
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Region     string
	TLS        bool
}

func CreateS3Client() (*S3Client, error) {
	s3Client := &S3Client{}
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	accessKey := os.Getenv("AWS_S3_ACCESSKEY")
	secretKey := os.Getenv("AWS_S3_SECRETKEY")
	region := os.Getenv("AWS_S3_REGION")
	useTLSStr := os.Getenv("AWS_S3_TLS")
	bucketName := os.Getenv("AWS_S3_BUCKET")
	skipVerifyStr := os.Getenv("AWS_S3_SKIPVERIFY")

	skipVerify := false
	if skipVerifyStr == "true" {
		skipVerify = true
	}

	useTLS := false
	if useTLSStr == "true" {
		useTLS = true
	}

	transport := http.DefaultTransport
	transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: skipVerify}

	mc, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:    useTLS,
		Region:    region,
		Transport: transport,
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.WithFields(log.Fields{"Endpoint": endpoint, "Region": region, "TLS": useTLS, "InsecureSkipVerify": skipVerify, "Bucket": bucketName}).Debug("Creating S3 client")

	s3Client.mc = mc
	s3Client.BucketName = bucketName
	s3Client.Endpoint = endpoint
	s3Client.TLS = useTLS
	s3Client.Region = region
	s3Client.AccessKey = accessKey
	s3Client.SecretKey = secretKey

	context := context.Background()
	bucket, err := mc.BucketExists(context, bucketName)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "BucketName": bucketName}).Error("Failed to check if bucket exists")
		return nil, err
	}
	if !bucket {
		err = mc.MakeBucket(context, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.WithFields(log.Fields{"Error": err, "Bucket": bucketName}).Error("Failed to create bucket")
			return nil, err
		}
		log.WithFields(log.Fields{"Bucket": bucketName}).Info("Creating bucket")
	}

	return s3Client, nil
}

type ProgressWriter struct {
	tracker *progress.Tracker
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	pw.tracker.Increment(int64(n))
	return n, nil
}

func (s3Client *S3Client) Upload(dir string, filename string, s3Filename string, filelength int64, tracker *progress.Tracker, quiet bool) error {
	f, err := os.Open(dir + "/" + filename)
	if err != nil {
		return err
	}

	progress := !quiet
	var reader io.Reader
	if progress {
		pw := &ProgressWriter{tracker: tracker}
		reader = io.TeeReader(bufio.NewReader(f), pw)
	} else {
		reader = bufio.NewReader(f)
	}

	_, err = s3Client.mc.PutObject(context.Background(), s3Client.BucketName, s3Filename, reader, filelength, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}

func (s3Client *S3Client) Download(filename string, s3Filename string, downloadDir string, tracker *progress.Tracker, quiet bool) error {
	file, err := s3Client.mc.GetObject(context.Background(), s3Client.BucketName, s3Filename, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	defer file.Close()

	destFile, err := os.Create(downloadDir + "/" + filename)
	if err != nil {
		return err

	}
	defer destFile.Close()

	progress := !quiet
	var writer io.Writer
	if progress {
		pw := &ProgressWriter{tracker: tracker}
		writer = io.MultiWriter(destFile, pw)
	} else {
		writer = destFile
	}

	_, err = io.Copy(writer, file)

	return err
}

func (s3Client *S3Client) Exists(filename string) bool {
	_, err := s3Client.mc.StatObject(context.Background(), s3Client.BucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return false
	}
	return true
}

func (s3Client *S3Client) Remove(filename string) error {
	return s3Client.mc.RemoveObject(context.Background(), s3Client.BucketName, filename, minio.RemoveObjectOptions{})
}
