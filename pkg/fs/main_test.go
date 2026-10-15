package fs

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("AWS_S3_ENDPOINT") == "" {
		fmt.Fprintln(os.Stderr, "pkg/fs tests need an S3 server (make starts3) and the AWS_S3_* variables, for example:")
		fmt.Fprintln(os.Stderr, "  export AWS_S3_ENDPOINT=localhost:9000 AWS_S3_ACCESSKEY=testaccesskey AWS_S3_SECRETKEY=testsecretkey AWS_S3_TLS=false AWS_S3_SKIPVERIFY=false AWS_S3_BUCKET=test")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
