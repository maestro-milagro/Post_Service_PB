package aws

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/maestro-milagro/Post_Service_PB/internal/lib/sl"
	"log/slog"
)

type AwsService struct {
	log    *slog.Logger
	Client *s3.Client
}

//type CrdnProvider struct {
//}
//
//func (c *CrdnProvider) Retrieve() (credentials.Value, error) {
//	return credentials.Value{
//		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
//		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
//		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
//		ProviderName:    "CrdnProvider",
//	}, nil
//}
//
//// IsExpired returns if the credentials are no longer valid, and need
//// to be retrieved.
//func (c *CrdnProvider) IsExpired() bool {
//
//}

func New(log *slog.Logger) *AwsService {
	// Создаем кастомный обработчик эндпоинтов, который для сервиса S3 и региона ru-central1 выдаст корректный URL
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == "ru-central1" {
			return aws.Endpoint{
				PartitionID:   "yc",
				URL:           "https://storage.yandexcloud.net",
				SigningRegion: "ru-central1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	// Подгружаем конфигурацию из ~/.aws/*
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		log.Error("failed to load configuration", sl.Err(err))

		return nil
	}

	// Создаем клиента для доступа к хранилищу S3
	client := s3.NewFromConfig(cfg)

	return &AwsService{
		log:    log,
		Client: client,
	}
}

func (a *AwsService) UploadFile(bucketName string, fileName string, largeObject []byte) error {
	//file, err := os.Open(fileName)
	//if err != nil {
	//	a.log.Error("Couldn't open file %v to upload. Here's why: %v\n", fileName, err)
	//} else {
	//		defer file.Close()
	largeBuffer := bytes.NewReader(largeObject)
	///		var partMiBs int64 = 10
	uploader := manager.NewUploader(a.Client)
	//	, func(u *manager.Uploader) {
	//		//			u.PartSize = partMiBs * 1024 * 1024
	//	}
	_, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   largeBuffer,
	})
	if err != nil {
		a.log.Error("Couldn't upload file %v to %v. Here's why: %v\n",
			fileName, bucketName, err)
	}
	//	}
	return err
}

func (a *AwsService) DownloadFile(bucketName string, filename string) ([]byte, error) {
	downloader := manager.NewDownloader(a.Client)
	buffer := manager.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(context.TODO(), buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		a.log.Error("Couldn't download large object from %v:%v. Here's why: %v\n",
			bucketName, filename, err)
	}
	return buffer.Bytes(), err
}

func (a *AwsService) DownloadList(bucketName string) ([]types.Object, error) {
	result, err := a.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	var contents []types.Object
	if err != nil {
		a.log.Error("Couldn't list objects in bucket %v. Here's why: %v\n", bucketName, err)
	} else {
		contents = result.Contents
	}
	return contents, err
}

func (a *AwsService) DeleteObjects(bucketName string, objectKeys []string) error {
	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	output, err := a.Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		a.log.Error("Couldn't delete objects from bucket %v. Here's why: %v\n", bucketName, err)
	} else {
		a.log.Info("Deleted %v objects.\n", len(output.Deleted))
	}
	return err
}
