package oss

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Store 覆盖 MinIO、AWS S3 及兼容 S3 API 的云厂商。
type S3Store struct {
	cfg    Config
	client *s3.S3
}

// NewS3 用已校验的 Config 创建 S3 兼容客户端。不探活存储，只保证 SDK session 能建出来。
func NewS3(cfg Config) (*S3Store, error) {
	cfg.applyDefaults()
	awsCfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Region:           aws.String(cfg.Region),
		S3ForcePathStyle: aws.Bool(cfg.PathStyle),
	}
	if cfg.Endpoint != "" {
		awsCfg.Endpoint = aws.String(cfg.Endpoint)
	}
	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, fmt.Errorf("oss: create s3 session: %w", err)
	}
	return &S3Store{cfg: cfg, client: s3.New(sess)}, nil
}

var (
	_ Store        = (*S3Store)(nil)
	_ PresignStore = (*S3Store)(nil)
)

func (s *S3Store) PutObject(ctx context.Context, in PutObjectInput) error {
	if in.Bucket == "" || in.Key == "" {
		return fmt.Errorf("oss: bucket and key are required")
	}
	if in.Body == nil {
		return fmt.Errorf("oss: body is nil")
	}
	if !s.cfg.AllowBucket(in.Bucket) {
		return fmt.Errorf("oss: bucket %s is not allowed", in.Bucket)
	}

	contentType := in.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	input := &s3.PutObjectInput{
		Bucket:      aws.String(in.Bucket),
		Key:         aws.String(strings.TrimPrefix(in.Key, "/")),
		Body:        in.Body,
		ContentType: aws.String(contentType),
	}
	if in.ContentLength > 0 {
		input.ContentLength = aws.Int64(in.ContentLength)
	}
	_, err := s.client.PutObjectWithContext(ctx, input)
	if err == nil {
		return nil
	}
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
		return fmt.Errorf("oss: upload cancelled: %w", err)
	}
	return fmt.Errorf("oss: put object: %w", err)
}

func (s *S3Store) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := s.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	})
	if err == nil {
		return true, nil
	}
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchKey, "NotFound":
			return false, nil
		}
	}
	return false, err
}

func (s *S3Store) ObjectURL(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	key = strings.TrimPrefix(key, "/")
	if bucket == "" || key == "" {
		return "", fmt.Errorf("oss: bucket and key are required")
	}
	if !s.cfg.AllowBucket(bucket) {
		return "", fmt.Errorf("oss: bucket %s is not allowed", bucket)
	}
	_ = ctx
	if s.cfg.IsPrivate(bucket) {
		if ttl <= 0 {
			ttl = s.cfg.AccessTTL(bucket)
		}
		req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		urlStr, err := req.Presign(ttl)
		if err != nil {
			return "", fmt.Errorf("oss: presign url: %w", err)
		}
		return urlStr, nil
	}
	urlStr := s.cfg.PublicURL(bucket, key)
	if urlStr == "" {
		return "", fmt.Errorf("oss: fileUrl is empty")
	}
	return urlStr, nil
}

func (s *S3Store) PresignPut(ctx context.Context, in PresignPutInput) (PresignPutResult, error) {
	if in.Bucket == "" || in.Key == "" {
		return PresignPutResult{}, fmt.Errorf("oss: bucket and key are required")
	}
	if !s.cfg.AllowBucket(in.Bucket) {
		return PresignPutResult{}, fmt.Errorf("oss: bucket %s is not allowed", in.Bucket)
	}
	contentType := in.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ttl := in.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	input := &s3.PutObjectInput{
		Bucket:      aws.String(in.Bucket),
		Key:         aws.String(strings.TrimPrefix(in.Key, "/")),
		ContentType: aws.String(contentType),
	}
	if in.ContentLength > 0 {
		input.ContentLength = aws.Int64(in.ContentLength)
	}
	req, _ := s.client.PutObjectRequest(input)
	if in.ContentLength > 0 {
		// Body 为空时 SDK 可能把长度写成 0，签发前写回真实长度，否则签不进 Content-Length。
		req.HTTPRequest.ContentLength = in.ContentLength
		req.HTTPRequest.Header.Set("Content-Length", strconv.FormatInt(in.ContentLength, 10))
	}
	urlStr, err := req.Presign(ttl)
	if err != nil {
		return PresignPutResult{}, fmt.Errorf("oss: presign put: %w", err)
	}
	_ = ctx
	headers := map[string]string{
		"Content-Type": contentType,
	}
	if in.ContentLength > 0 {
		headers["Content-Length"] = strconv.FormatInt(in.ContentLength, 10)
	}
	return PresignPutResult{
		URL:     urlStr,
		Method:  "PUT",
		Headers: headers,
	}, nil
}
func (s *S3Store) HeadObject(ctx context.Context, bucket, key string) (ObjectHead, error) {
	if bucket == "" || key == "" {
		return ObjectHead{}, fmt.Errorf("oss: bucket and key are required")
	}
	out, err := s.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey, "NotFound", "NoSuchBucket":
				return ObjectHead{}, fmt.Errorf("%w", ErrNotFound)
			}
			if rf, ok := err.(awserr.RequestFailure); ok && rf.StatusCode() == http.StatusNotFound {
				return ObjectHead{}, fmt.Errorf("%w", ErrNotFound)
			}
		}
		return ObjectHead{}, err
	}
	head := ObjectHead{}
	if out.ContentLength != nil {
		head.Size = *out.ContentLength
	}
	if out.ContentType != nil {
		head.ContentType = *out.ContentType
	}
	return head, nil
}

func (s *S3Store) DeleteObject(ctx context.Context, bucket, key string) error {
	if bucket == "" || key == "" {
		return fmt.Errorf("oss: bucket and key are required")
	}
	if !s.cfg.AllowBucket(bucket) {
		return fmt.Errorf("oss: bucket %s is not allowed", bucket)
	}
	_, err := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	})
	if err != nil {
		return fmt.Errorf("oss: delete object: %w", err)
	}
	return nil
}
