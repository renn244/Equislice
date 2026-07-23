package handler

import (
	"backend/internal/dto"
	"backend/internal/services"
	"backend/internal/util"
	"errors"
	"net/http"
	"strconv"
)

type PanoramaHandler struct {
	PanoramaService *services.PanoramaService
}

func NewPanoramaHandler(panoramaService *services.PanoramaService) *PanoramaHandler {
	return &PanoramaHandler{
		PanoramaService: panoramaService,
	}
}

func (h *PanoramaHandler) PostPanorama(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 50 << 20 // 50 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err := r.ParseMultipartForm(32 << 20) // 32 MB
	if err != nil {
		_, isMaxByte := errors.AsType[*http.MaxBytesError](err)
		if isMaxByte {
			util.WriteError(w, http.StatusRequestEntityTooLarge, "uploaded file is too large")
			return
		}

		util.WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	row := r.FormValue("rows")
	rowInt, err := strconv.Atoi(row)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid rows")
		return
	}
	if rowInt <= 0 {
		util.WriteError(w, http.StatusBadRequest, "invalid rows")
		return
	}

	column := r.FormValue("columns")
	columnInt, err := strconv.Atoi(column)
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid columns")
		return
	}
	if columnInt <= 0 {
		util.WriteError(w, http.StatusBadRequest, "invalid columns")
		return
	}

	fileFormat := r.FormValue("fileFormat")

	file, header, err := r.FormFile("file")
	if err != nil {
		util.WriteError(w, http.StatusBadRequest, "invalid file")
		return
	}
	defer file.Close()

	if header.Header.Get("Content-Type") != "image/jpeg" && header.Header.Get("Content-Type") != "image/png" {
		util.WriteError(w, http.StatusBadRequest, "invalid image format")
		return
	}

	jobId, err := h.PanoramaService.PostPanorama(
		r.Context(),
		header.Filename,
		file,
		rowInt,
		columnInt,
		fileFormat,
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
		Status: data.Status,
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
