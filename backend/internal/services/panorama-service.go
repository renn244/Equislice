package services

import (
	"archive/zip"
	"backend/internal/dto"
	"backend/internal/ports"
	"backend/internal/util/constants"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PanoramaService struct {
	PanoramaStorage      ports.FileService
	PanoramaSliceStorage ports.FileService
	DB                   ports.DBService[dto.PanoramaEntity]
	Queue                ports.QueueService
}

func NewPanoramaService(
	panoramaStorage ports.FileService, panoramaSliceStorage ports.FileService, db ports.DBService[dto.PanoramaEntity], queue ports.QueueService,
) *PanoramaService {
	return &PanoramaService{
		PanoramaStorage:      panoramaStorage,
		PanoramaSliceStorage: panoramaSliceStorage,
		DB:                   db,
		Queue:                queue,
	}
}

func (s *PanoramaService) PostPanorama(ctx context.Context, panoramaFileName string, row int, column int, fileFormats string) (string, error) {
	v4, err := uuid.NewRandom()
	if err != nil {
		return "", errors.New("failed to generate v4 uuid")
	}

	jobId := v4.String()

	_, err = s.DB.Insert(ctx, dto.PanoramaEntity{
		JobId:             jobId,
		InitialPanoramaId: panoramaFileName,
		Status:            "Queued",
		Row:               row,
		Column:            column,
		FileFormat:        fileFormats,
	}, &jobId)
	if err != nil {
		return "", errors.New("failed to add entity")
	}

	JobQueueMessage, err := json.Marshal(dto.PanoramaQueueMessage{
		JobId: jobId,
	})
	if err != nil {
		s.DB.Delete(ctx, jobId)
		return "", errors.New("failed to add queue")
	}

	_, err = s.Queue.Enqueue(ctx, string(JobQueueMessage))
	if err != nil {
		s.DB.Delete(ctx, jobId)
		return "", errors.New("failed to enqueue message")
	}

	return jobId, nil
}

func (s *PanoramaService) GetStatus(ctx context.Context, jobId string) (*dto.PanoramaEntity, error) {
	body, err := s.DB.Find(ctx, jobId)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func (s *PanoramaService) GetUploadUrl(ctx context.Context, fileName string, contentType string) (string, string, error) {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return "", "", errors.New("file extension required")
	}
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	if !slices.Contains(constants.ValidContentTypeList, contentType) {
		return "", "", errors.New("invalid content type")
	}
	if ext != "jpg" && ext != "jpeg" && ext != "png" {
		return "", "", errors.New("invalid content type")
	}

	v4, err := uuid.NewRandom()
	if err != nil {
		return "", "", errors.New("failed to generate v4 uuid")
	}

	blobName := v4.String() + "-" + fileName
	duration := time.Hour * 1

	uploadUrl, err := s.PanoramaStorage.GenerateUploadUrl(ctx, blobName, contentType, duration)
	if err != nil {
		return "", "", errors.New("failed to generate upload url")
	}

	return uploadUrl, blobName, nil
}

func (s *PanoramaService) GetShareUrl(ctx context.Context, jobId string) ([]string, error) {
	body, err := s.DB.Find(ctx, jobId)
	if err != nil {
		return nil, errors.New("failed to find job")
	}

	if body.Status != "Completed" {
		return nil, errors.New("not yet complete")
	}

	if body.PanoramaSliceId == nil {
		return nil, errors.New("not yet complete")
	}

	var sasUrls []string

	for i := range body.Row {
		for j := range body.Column {
			blobName := *body.PanoramaSliceId + "\\" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".jpg"

			var sasUrl string
			sasUrl, err = s.PanoramaSliceStorage.GenerateShareUrl(ctx, blobName, 1*time.Hour)
			if err != nil {
				return nil, err
			}

			sasUrls = append(sasUrls, sasUrl)
		}
	}

	return sasUrls, nil
}

func (s *PanoramaService) GetArchive(ctx context.Context, jobId string) ([]byte, error) {
	body, err := s.DB.Find(ctx, jobId)
	if err != nil {
		return nil, errors.New("failed to find job")
	}

	if body.Status != "Completed" || body.PanoramaSliceId == nil {
		return nil, errors.New("not yet complete")
	}

	var archive bytes.Buffer
	zipWriter := zip.NewWriter(&archive)

	for i := range body.Row {
		for j := range body.Column {
			blobName := *body.PanoramaSliceId + "\\" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".jpg"
			tileName := fmt.Sprintf("tile_r%02d_c%02d.jpg", i+1, j+1)

			tileWriter, createErr := zipWriter.Create(tileName)
			if createErr != nil {
				return nil, createErr
			}

			if downloadErr := s.PanoramaSliceStorage.Download(ctx, blobName, tileWriter); downloadErr != nil {
				return nil, downloadErr
			}
		}
	}

	if err = zipWriter.Close(); err != nil {
		return nil, err
	}

	return archive.Bytes(), nil
}
