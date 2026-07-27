package handler

import (
	"backend/internal/dto"
	"backend/internal/services"
	"backend/internal/util"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type PanoramaHandler struct {
	PanoramaService *services.PanoramaService
}

func NewPanoramaHandler(panoramaService *services.PanoramaService) *PanoramaHandler {
	return &PanoramaHandler{
		PanoramaService: panoramaService,
	}
}

func (h *PanoramaHandler) PostPanorama(w http.ResponseWriter, r *http.Request) {
	var body dto.PostPanoramaRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		util.WriteError(w, http.StatusUnprocessableEntity, "Failed to decode request body")
		return
	}

	err = validate.Struct(&body)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	jobId, err := h.PanoramaService.PostPanorama(
		r.Context(),
		body.File,
		body.Rows,
		body.Columns,
		body.FileFormats,
	)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, "error post panorama slice")
		return
	}

	util.WriteJson[dto.PostPanoramaResponse](w, http.StatusCreated, dto.PostPanoramaResponse{
		JobId: jobId,
	})
}

func (h *PanoramaHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("job-id")
	if jobId == "" {
		util.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	data, err := h.PanoramaService.GetStatus(r.Context(), jobId)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, "error getting job status")
		return
	}

	util.WriteJson[dto.GetStatusPanoramaResponse](w, http.StatusOK, dto.GetStatusPanoramaResponse{
		Status:  data.Status,
		Columns: data.Column,
		Rows:    data.Row,
	})
}

func (h *PanoramaHandler) GetUploadUrl(w http.ResponseWriter, r *http.Request) {
	var body dto.GetUploadUrlRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		util.WriteError(w, http.StatusUnprocessableEntity, "invalid request body")
		return
	}

	err = validate.Struct(body)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	uploadUrl, blobName, err := h.PanoramaService.GetUploadUrl(r.Context(), body.FileName, body.ContentType)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, "error getting upload url")
		return
	}

	util.WriteJson[dto.GetUploadUrlResponse](w, http.StatusOK, dto.GetUploadUrlResponse{
		UrlSAS:   uploadUrl,
		BlobName: blobName,
	})
}

func (h *PanoramaHandler) GetShareUrl(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("job-id")
	if jobId == "" {
		util.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	sasUrls, err := h.PanoramaService.GetShareUrl(r.Context(), jobId)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, "error getting share url")
		return
	}

	util.WriteJson[dto.GetSASUrlResponse](w, http.StatusOK, dto.GetSASUrlResponse{
		UrlsSAS: sasUrls,
	})
}

func (h *PanoramaHandler) GetArchive(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("job-id")
	if jobId == "" {
		util.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	archive, err := h.PanoramaService.GetArchive(r.Context(), jobId)
	if err != nil {
		util.WriteError(w, http.StatusInternalServerError, "error creating tile archive")
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="equislice-`+jobId+`.zip"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(archive)
}
