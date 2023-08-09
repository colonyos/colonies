package fs

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/k0kubun/go-ansi"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/schollz/progressbar/v3"
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
		log.Fatal(err)
		return nil, err
	}

	log.WithFields(log.Fields{"Endpoint": endpoint, "AccessKey": accessKey, "SecretKey": secretKey, "Region": region, "TLS": useTLS, "InsecureSkipVerify": skipVerify, "Bucket": bucketName}).Debug("Creating S3 client")

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

	progress := true
	var reader io.Reader
	if progress {
		bar := progressbar.NewOptions(int(filelength),
			progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription("[cyan] Uploading "+filename),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[blue]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))

		reader = io.TeeReader(bufio.NewReader(f), bar)
	} else {
		reader = bufio.NewReader(f)
	}

	fmt.Println("FileLength", filelength)

	_, err = s3Client.mc.PutObject(context.Background(), s3Client.BucketName, filename, reader, filelength, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}

func (s3Client *S3Client) Download(filename string, downloadDir string) error {
	file, err := s3Client.mc.GetObject(context.Background(), s3Client.BucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	defer file.Close()

	destFile, err := os.Create(downloadDir + "/" + filename)
	if err != nil {
		return err

	}
	defer destFile.Close()

	progress := true
	var writer io.Writer
	if progress {
		objInfo, err := s3Client.mc.StatObject(context.Background(), s3Client.BucketName, filename, minio.StatObjectOptions{})
		if err != nil {
			return err
		}

		bar := progressbar.NewOptions(int(objInfo.Size),
			progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription("[cyan] Downloading "+filename),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[blue]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))

		writer = io.MultiWriter(destFile, bar)
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
