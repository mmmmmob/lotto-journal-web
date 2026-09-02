package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/image/draw"
)



type StorageService struct {
	client     *s3.Client
	bucketName string
}

func NewStorageService(accountID, accessKey, secretKey, bucketName string) (*StorageService, error) {
	if accountID == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing required storage configuration: accountID, accessKey, secretKey, or bucketName")
	}

	// Cloudflare R2 endpoint format: https://<account_id>.r2.cloudflarestorage.com
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 credentials: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
	})

	return &StorageService{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadAndResizeImage decodes, downscales to maximum 1280px (maintaining aspect ratio),
// encodes to JPEG, uploads to Cloudflare R2, and returns the image as a base64 Data URL.
func (s *StorageService) UploadAndResizeImage(ctx context.Context, imageReader io.Reader, storageKey string) (string, error) {
	// 1. Decode image
	img, _, err := image.Decode(imageReader)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Resize maintaining aspect ratio if any dimension exceeds 1280px
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	maxDim := 1280

	if w > maxDim || h > maxDim {
		var newW, newH int
		if w > h {
			newW = maxDim
			newH = (h * maxDim) / w
		} else {
			newH = maxDim
			newW = (w * maxDim) / h
		}
		dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
		img = dst
	}

	// 3. Encode back to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf("failed to encode image to jpeg: %w", err)
	}

	imageBytes := buf.Bytes()

	// 4. Upload to Cloudflare R2
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(storageKey),
		Body:        bytes.NewReader(imageBytes),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to R2 storage: %w", err)
	}

	// 5. Generate base64 data URL
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Data)

	return dataURL, nil
}
