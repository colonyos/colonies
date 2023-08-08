package fs

// func GenerateRandomID() string {
// 	uuid := uuid.New()
// 	crypto := crypto.CreateCrypto()
// 	return crypto.GenerateHash(uuid.String())[0:10]
// }

// func TestCreateMinio(t *testing.T) {
// 	endpoint := os.Getenv("AWS_S3_ENDPOINT_TEST")
// 	accessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_TEST")
// 	secretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY_TEST")
// 	useSSL := false

// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	assert.Nil(t, err)
// 	log.Printf("%#v\n", minioClient)
// }

// func TestSaveFile(t *testing.T) {
// 	endpoint := os.Getenv("AWS_S3_ENDPOINT_TEST")
// 	accessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_TEST")
// 	secretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY_TEST")
// 	useSSL := false

// 	bucketName := GenerateRandomID()
// 	s3, err := s3.New(endpoint, "", accessKeyID, secretAccessKey, useSSL, true, bucketName, "")
// 	assert.Nil(t, err)
// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	assert.Nil(t, err)

// 	fileContents := []byte("testdata")

// 	err = s3.SaveFile("dummydata", "tmp/", fileContents)
// 	assert.Nil(t, err)

// 	_, err = minioClient.StatObject(context.Background(), bucketName, "/tmp/dummydata", minio.StatObjectOptions{})
// 	assert.Nil(t, err)

// 	minioClient.RemoveObject(context.Background(), bucketName, "/tmp/dummydata", minio.RemoveObjectOptions{})
// 	minioClient.RemoveBucket(context.Background(), bucketName)
// }

// func TestGetFile(t *testing.T) {
// 	endpoint := os.Getenv("AWS_S3_ENDPOINT_TEST")
// 	accessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_TEST")
// 	secretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY_TEST")
// 	useSSL := false

// 	bucketName := GenerateRandomID()
// 	s3, err := s3.New(endpoint, "", accessKeyID, secretAccessKey, useSSL, true, bucketName, "")
// 	assert.Nil(t, err)
// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	assert.Nil(t, err)

// 	fileContents := []byte("testdata")

// 	_, err = minioClient.PutObject(context.Background(), bucketName, "tmp/dummydata", bytes.NewReader(fileContents), int64(len(fileContents)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
// 	assert.Nil(t, err)

// 	f, err := s3.GetFile("/tmp/dummydata")
// 	assert.Nil(t, err)
// 	assert.Equal(t, fileContents, *f)

// 	minioClient.RemoveObject(context.Background(), bucketName, "tmp/dummydata", minio.RemoveObjectOptions{})
// 	minioClient.RemoveBucket(context.Background(), bucketName)
// }

// func TestExists(t *testing.T) {
// 	endpoint := os.Getenv("AWS_S3_ENDPOINT_TEST")
// 	accessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_TEST")
// 	secretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY_TEST")
// 	useSSL := false

// 	bucketName := GenerateRandomID()
// 	s3, err := s3.New(endpoint, "", accessKeyID, secretAccessKey, useSSL, true, bucketName, "")
// 	assert.Nil(t, err)
// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	assert.Nil(t, err)

// 	fileContents := []byte("testdata")

// 	_, err = minioClient.PutObject(context.Background(), bucketName, "tmp/dummydata", bytes.NewReader(fileContents), int64(len(fileContents)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
// 	assert.Nil(t, err)

// 	exists := s3.FileExists("/tmp/dummydata")

// 	assert.True(t, exists)
// 	minioClient.RemoveObject(context.Background(), bucketName, "tmp/dummydata", minio.RemoveObjectOptions{})
// 	minioClient.RemoveBucket(context.Background(), bucketName)
// }

// func TestRemoveFile(t *testing.T) {
// 	endpoint := os.Getenv("AWS_S3_ENDPOINT_TEST")
// 	accessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_TEST")
// 	secretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY_TEST")
// 	useSSL := false

// 	bucketName := GenerateRandomID()
// 	s3, err := s3.New(endpoint, "", accessKeyID, secretAccessKey, useSSL, true, bucketName, "")
// 	assert.Nil(t, err)
// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	assert.Nil(t, err)

// 	fileContents := []byte("testdata")

// 	_, err = minioClient.PutObject(context.Background(), bucketName, "tmp/dummydata", bytes.NewReader(fileContents), int64(len(fileContents)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
// 	assert.Nil(t, err)

// 	err = s3.RemoveFile("tmp/dummydata")
// 	assert.Nil(t, err)

// 	err = minioClient.RemoveBucket(context.Background(), bucketName)
// 	assert.Nil(t, err)
// }

// func TestRemoveDirectory(t *testing.T) {
// 	endpoint := os.Getenv("AWS_S3_ENDPOINT_TEST")
// 	accessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_TEST")
// 	secretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY_TEST")
// 	useSSL := false

// 	bucketName := GenerateRandomID()
// 	s3, err := s3.New(endpoint, "", accessKeyID, secretAccessKey, useSSL, true, bucketName, "")
// 	assert.Nil(t, err)
// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	assert.Nil(t, err)

// 	fileContents := []byte("testdata")

// 	_, err = minioClient.PutObject(context.Background(), bucketName, "tmp/dummydata", bytes.NewReader(fileContents), int64(len(fileContents)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
// 	assert.Nil(t, err)

// 	err = s3.RemoveDirectory("tmp/")
// 	assert.Nil(t, err)

// 	err = minioClient.RemoveBucket(context.Background(), bucketName)
// 	assert.Nil(t, err)
// }

// func TestListDirectory(t *testing.T) {
// 	endpoint := os.Getenv("AWS_S3_ENDPOINT_TEST")
// 	accessKeyID := os.Getenv("AWS_S3_ACCESS_KEY_TEST")
// 	secretAccessKey := os.Getenv("AWS_S3_SECRET_ACCESS_KEY_TEST")
// 	useSSL := false

// 	bucketName := GenerateRandomID()
// 	fmt.Println(bucketName)
// 	s3, err := s3.New(endpoint, "", accessKeyID, secretAccessKey, useSSL, true, bucketName, "")
// 	assert.Nil(t, err)
// 	minioClient, err := minio.New(endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
// 		Secure: useSSL,
// 	})
// 	assert.Nil(t, err)

// 	for k := 0; k < 10; k++ {
// 		fileContents := []byte("testdata" + strconv.Itoa(k))
// 		_, err = minioClient.PutObject(context.Background(), bucketName, "tmp/temp2/dummydata_"+strconv.Itoa(k), bytes.NewReader(fileContents), int64(len(fileContents)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
// 		assert.Nil(t, err)
// 	}

// 	objectCh := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{Prefix: "/tmp/temp2/"})
// 	keys := []string{}
// 	for object := range objectCh {
// 		if object.Err != nil {
// 			fmt.Println(object.Err)
// 			return
// 		}
// 		fmt.Println(object.Key)
// 		keys = append(keys, object.Key)
// 	}

// 	err = s3.RemoveDirectory("tmp/temp2/")
// 	assert.Nil(t, err)

// 	err = s3.RemoveDirectory("tmp/")
// 	assert.Nil(t, err)

// 	err = minioClient.RemoveBucket(context.Background(), bucketName)
// 	assert.Nil(t, err)
// }
