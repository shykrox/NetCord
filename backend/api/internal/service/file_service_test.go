package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"
	"netcord/backend/api/internal/storage"

	"github.com/google/uuid"
)

func TestFileServiceUploadDetectsMimeAndStoresAttachment(t *testing.T) {
	files := newFakeFileRepository()
	store := newFakeObjectStore()
	service := NewFileService(files, store, "attachments", 1024)
	userID := uuid.New()

	attachment, err := service.Upload(context.Background(), userID, UploadFileInput{
		OriginalFilename: "../hello.txt",
		SizeBytes:        int64(len("hello world")),
		Reader:           strings.NewReader("hello world"),
	})
	if err != nil {
		t.Fatalf("upload file: %v", err)
	}

	if attachment.OriginalFilename != "hello.txt" {
		t.Fatalf("expected sanitized filename, got %q", attachment.OriginalFilename)
	}
	if attachment.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("expected detected text mime, got %q", attachment.ContentType)
	}
	if attachment.DownloadURL != "/files/"+attachment.ID.String() {
		t.Fatalf("unexpected download url: %s", attachment.DownloadURL)
	}

	stored := files.attachments[attachment.ID]
	if stored.ObjectKey == "" || !strings.HasPrefix(stored.ObjectKey, "attachments/") {
		t.Fatalf("unexpected object key: %q", stored.ObjectKey)
	}
	if store.objects[stored.ObjectKey] != "hello world" {
		t.Fatalf("stored object mismatch: %q", store.objects[stored.ObjectKey])
	}
}

func TestFileServiceUploadRejectsOversizedFiles(t *testing.T) {
	service := NewFileService(newFakeFileRepository(), newFakeObjectStore(), "attachments", 3)

	_, err := service.Upload(context.Background(), uuid.New(), UploadFileInput{
		OriginalFilename: "hello.txt",
		SizeBytes:        5,
		Reader:           strings.NewReader("hello"),
	})
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestFileServiceDownloadRequiresAccessibleAttachment(t *testing.T) {
	files := newFakeFileRepository()
	store := newFakeObjectStore()
	service := NewFileService(files, store, "attachments", 1024)
	userID := uuid.New()
	attachmentID := uuid.New()
	files.attachments[attachmentID] = models.MessageAttachment{
		ID:               attachmentID,
		UploaderID:       userID,
		Bucket:           "attachments",
		ObjectKey:        "attachments/random/file",
		OriginalFilename: "hello.txt",
		ContentType:      "text/plain; charset=utf-8",
		SizeBytes:        5,
	}
	store.objects["attachments/random/file"] = "hello"

	download, err := service.Download(context.Background(), userID, attachmentID)
	if err != nil {
		t.Fatalf("download file: %v", err)
	}
	defer download.Body.Close()

	body, err := io.ReadAll(download.Body)
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if string(body) != "hello" {
		t.Fatalf("expected hello, got %q", body)
	}
}

type fakeFileRepository struct {
	attachments map[uuid.UUID]models.MessageAttachment
}

func newFakeFileRepository() *fakeFileRepository {
	return &fakeFileRepository{attachments: make(map[uuid.UUID]models.MessageAttachment)}
}

func (r *fakeFileRepository) CreateAttachment(ctx context.Context, attachment models.MessageAttachment) (models.MessageAttachment, error) {
	r.attachments[attachment.ID] = attachment
	return attachment, nil
}

func (r *fakeFileRepository) GetAttachmentForUser(ctx context.Context, attachmentID, userID uuid.UUID) (models.MessageAttachment, error) {
	attachment, ok := r.attachments[attachmentID]
	if !ok || attachment.UploaderID != userID {
		return models.MessageAttachment{}, repository.ErrAttachmentNotFound
	}
	return attachment, nil
}

type fakeObjectStore struct {
	objects map[string]string
}

func newFakeObjectStore() *fakeObjectStore {
	return &fakeObjectStore{objects: make(map[string]string)}
}

func (s *fakeObjectStore) EnsureBucket(ctx context.Context, bucket string) error {
	return nil
}

func (s *fakeObjectStore) Put(ctx context.Context, bucket, objectKey string, reader io.Reader, size int64, contentType string) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s.objects[objectKey] = string(body)
	return nil
}

func (s *fakeObjectStore) Get(ctx context.Context, bucket, objectKey string) (io.ReadCloser, storage.ObjectInfo, error) {
	body, ok := s.objects[objectKey]
	if !ok {
		return nil, storage.ObjectInfo{}, repository.ErrAttachmentNotFound
	}
	return io.NopCloser(bytes.NewBufferString(body)), storage.ObjectInfo{
		Size:        int64(len(body)),
		ContentType: "text/plain; charset=utf-8",
	}, nil
}
