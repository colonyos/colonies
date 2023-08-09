package fs

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type S3Client struct {
	mc         *minio.Client
	bucketName string
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
		log.Fatal(err)
		return nil, err
	}

	log.WithFields(log.Fields{"Endpoint": endpoint, "AccessKey": accessKey, "SecretKey": secretKey, "Region": region, "TLS": useTLS, "InsecureSkipVerify": skipVerify, "Bucket": bucketName}).Debug("Creating S3 client")

	s3Client.mc = mc
	s3Client.bucketName = bucketName

	context := context.Background()
	bucket, err := mc.BucketExists(context, bucketName)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "Bucket": bucket, "BucketName": bucketName}).Fatal("Failed to check if bucket exists")
		return nil, err
	}
	if !bucket {
		err = mc.MakeBucket(context, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.WithFields(log.Fields{"Error": err, "Bucket": bucketName}).Fatal("Failed to create bucket")
			return nil, err
		}
		log.WithFields(log.Fields{"Bucket": bucketName}).Info("Creating bucket")
	}

	return s3Client, nil
}

func (s3Client *S3Client) Upload(dir string, filename string, filelength int64) error {
	f, err := os.Open(dir + "/" + filename)
	if err != nil {
		return err
	}

	_, err = s3Client.mc.PutObject(context.Background(), s3Client.bucketName, filename, bufio.NewReader(f), filelength, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}

func (s3Client *S3Client) Download(filename string, downloadDir string) error {
	file, err := s3Client.mc.GetObject(context.Background(), s3Client.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	defer file.Close()

	destFile, err := os.Create(downloadDir + "/" + filename)
	if err != nil {
		return err

	}
	defer destFile.Close()

	_, err = io.Copy(destFile, file)

	return err
}

func (s3Client *S3Client) Exists(filename string) bool {
	_, err := s3Client.mc.StatObject(context.Background(), s3Client.bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return false
	}
	return true
}

func (s3Client *S3Client) Remove(filename string) error {
	return s3Client.mc.RemoveObject(context.Background(), s3Client.bucketName, filename, minio.RemoveObjectOptions{})
}
