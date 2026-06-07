package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"
	"netcord/backend/api/internal/storage"

	"github.com/google/uuid"
)

const sniffBytes = 512

type UploadFileInput struct {
	OriginalFilename string
	SizeBytes        int64
	Reader           io.Reader
}

type FileDownload struct {
	Attachment  models.PublicAttachment
	Body        io.ReadCloser
	ContentType string
	SizeBytes   int64
}

type FileService struct {
	files          repository.FileRepository
	store          storage.ObjectStore
	bucket         string
	maxUploadBytes int64
}

func NewFileService(files repository.FileRepository, store storage.ObjectStore, bucket string, maxUploadBytes int64) *FileService {
	return &FileService{
		files:          files,
		store:          store,
		bucket:         strings.TrimSpace(bucket),
		maxUploadBytes: maxUploadBytes,
	}
}

func (s *FileService) MaxUploadBytes() int64 {
	return s.maxUploadBytes
}

func (s *FileService) Upload(ctx context.Context, userID uuid.UUID, input UploadFileInput) (models.PublicAttachment, error) {
	if s.files == nil || s.store == nil || s.bucket == "" {
		return models.PublicAttachment{}, errors.New("file service is not configured")
	}
	if input.Reader == nil {
		return models.PublicAttachment{}, &ValidationError{Fields: map[string]string{"file": "file is required"}}
	}
	if input.SizeBytes <= 0 {
		return models.PublicAttachment{}, &ValidationError{Fields: map[string]string{"file": "file must not be empty"}}
	}
	if s.maxUploadBytes > 0 && input.SizeBytes > s.maxUploadBytes {
		return models.PublicAttachment{}, &ValidationError{Fields: map[string]string{"file": "file exceeds maximum upload size"}}
	}

	filename := cleanFilename(input.OriginalFilename)
	contentType, reader, err := detectContentType(input.Reader)
	if err != nil {
		return models.PublicAttachment{}, err
	}

	attachment := models.MessageAttachment{
		ID:               uuid.New(),
		UploaderID:       userID,
		Bucket:           s.bucket,
		ObjectKey:        newAttachmentObjectKey(filename),
		OriginalFilename: filename,
		ContentType:      contentType,
		SizeBytes:        input.SizeBytes,
	}

	if err := s.store.Put(ctx, attachment.Bucket, attachment.ObjectKey, reader, attachment.SizeBytes, attachment.ContentType); err != nil {
		return models.PublicAttachment{}, err
	}

	created, err := s.files.CreateAttachment(ctx, attachment)
	if err != nil {
		return models.PublicAttachment{}, err
	}

	return created.Public(), nil
}

func (s *FileService) Download(ctx context.Context, userID, attachmentID uuid.UUID) (FileDownload, error) {
	if s.files == nil || s.store == nil {
		return FileDownload{}, errors.New("file service is not configured")
	}

	attachment, err := s.files.GetAttachmentForUser(ctx, attachmentID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAttachmentNotFound) {
			return FileDownload{}, ErrNotFound
		}
		return FileDownload{}, err
	}

	body, info, err := s.store.Get(ctx, attachment.Bucket, attachment.ObjectKey)
	if err != nil {
		return FileDownload{}, err
	}

	contentType := attachment.ContentType
	if info.ContentType != "" {
		contentType = info.ContentType
	}

	sizeBytes := attachment.SizeBytes
	if info.Size > 0 {
		sizeBytes = info.Size
	}

	return FileDownload{
		Attachment:  attachment.Public(),
		Body:        body,
		ContentType: contentType,
		SizeBytes:   sizeBytes,
	}, nil
}

func detectContentType(reader io.Reader) (string, io.Reader, error) {
	header := make([]byte, sniffBytes)
	n, err := io.ReadFull(reader, header)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", nil, err
	}
	header = header[:n]

	contentType := http.DetectContentType(header)
	return contentType, io.MultiReader(bytes.NewReader(header), reader), nil
}

func cleanFilename(filename string) string {
	filename = strings.TrimSpace(filepath.Base(filename))
	if filename == "" || filename == "." {
		return "upload"
	}
	return filename
}

func newAttachmentObjectKey(filename string) string {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("attachments/%s/%s", uuid.NewString(), cleanObjectSuffix(filename))
	}

	return fmt.Sprintf("attachments/%s/%s/%s", uuid.NewString(), hex.EncodeToString(randomBytes), cleanObjectSuffix(filename))
}

func cleanObjectSuffix(filename string) string {
	extension := strings.ToLower(filepath.Ext(filename))
	if extension == "" || len(extension) > 16 {
		return "file"
	}
	return strings.TrimPrefix(extension, ".")
}
