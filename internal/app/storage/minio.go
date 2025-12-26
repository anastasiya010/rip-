package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
	publicURL  string
}

func NewMinioClient() (*MinioClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	if accessKeyID == "" {
		accessKeyID = "minioadmin"
	}

	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	if secretAccessKey == "" {
		secretAccessKey = "minioadmin"
	}

	useSSL := os.Getenv("MINIO_USE_SSL")
	useSSLBool := useSSL == "true"

	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "medications"
	}

	publicURL := os.Getenv("MINIO_PUBLIC_BASE")
	if publicURL == "" {
		publicURL = fmt.Sprintf("http://%s/%s", endpoint, bucketName)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSLBool,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	mc := &MinioClient{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}

	// Создаем bucket если его нет
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		logrus.Infof("Bucket %s created successfully", bucketName)
	}

	// Устанавливаем публичный доступ для чтения
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, bucketName)

	err = client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		logrus.Warnf("Failed to set bucket policy (non-critical): %v", err)
	}

	return mc, nil
}

// UploadImage загружает изображение в Minio и возвращает ключ (путь к файлу)
func (mc *MinioClient) UploadImage(ctx context.Context, file io.Reader, contentType string) (string, error) {
	// Генерируем уникальное имя файла на латинице
	ext := ""
	if contentType != "" {
		parts := strings.Split(contentType, "/")
		if len(parts) > 1 {
			ext = "." + parts[1]
		}
	}
	if ext == "" {
		ext = ".jpg" // по умолчанию
	}

	// Генерируем имя на латинице
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	objectName := fmt.Sprintf("images/%s", fileName)

	// Загружаем файл
	_, err := mc.client.PutObject(ctx, mc.bucketName, objectName, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return objectName, nil
}

// DeleteImage удаляет изображение из Minio
func (mc *MinioClient) DeleteImage(ctx context.Context, objectName string) error {
	if objectName == "" {
		return nil // Нет изображения для удаления
	}

	err := mc.client.RemoveObject(ctx, mc.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// GetImageURL возвращает публичный URL изображения
func (mc *MinioClient) GetImageURL(objectName string) string {
	if objectName == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(mc.publicURL, "/"), objectName)
}

// GeneratePresignedURL генерирует временную ссылку для доступа к изображению
func (mc *MinioClient) GeneratePresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if objectName == "" {
		return "", fmt.Errorf("object name is empty")
	}

	url, err := mc.client.PresignedGetObject(ctx, mc.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

