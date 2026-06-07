package httpapi

import (
	"io"
	"mime"
	"net/http"
	"strconv"

	"netcord/backend/api/internal/service"
)

const multipartOverheadBytes = 1 << 20

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	if s.fileService == nil {
		writeError(w, http.StatusInternalServerError, "files_unavailable", "file service is unavailable", nil)
		return
	}

	maxUploadBytes := s.fileService.MaxUploadBytes()
	if maxUploadBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+multipartOverheadBytes)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_multipart", "multipart field file is required", nil)
		return
	}
	defer file.Close()

	attachment, err := s.fileService.Upload(r.Context(), userID, service.UploadFileInput{
		OriginalFilename: header.Filename,
		SizeBytes:        header.Size,
		Reader:           file,
	})
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, attachment)
}

func (s *Server) getFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	if s.fileService == nil {
		writeError(w, http.StatusInternalServerError, "files_unavailable", "file service is unavailable", nil)
		return
	}

	fileID, ok := pathUUID(w, r, "file_id")
	if !ok {
		return
	}

	download, err := s.fileService.Download(r.Context(), userID, fileID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	defer download.Body.Close()

	w.Header().Set("Content-Type", download.ContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": download.Attachment.OriginalFilename,
	}))
	if download.SizeBytes > 0 {
		w.Header().Set("Content-Length", int64String(download.SizeBytes))
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, download.Body)
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}
