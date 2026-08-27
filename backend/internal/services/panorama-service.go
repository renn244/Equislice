package services

import (
	"archive/zip"
	"backend/internal/dto"
	"backend/internal/ports"
	"backend/internal/util"
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

var (
	ErrFileExtensionRequired = errors.New("file extension required")
	ErrInvalidContentType    = errors.New("invalid content type")
	ErrJobNotComplete        = errors.New("job not complete")
)

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

func (s *PanoramaService) PostPanorama(ctx context.Context, panoramaFileName string, row int, column int, fileFormats string, fileNameFormat string) (string, error) {
	v4, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate v4 uuid: %w", err)
	}

	jobId := v4.String()

	_, err = s.DB.Insert(ctx, dto.PanoramaEntity{
		JobId:             jobId,
		InitialPanoramaId: panoramaFileName,
		Status:            "Queued",
		Row:               row,
		Column:            column,
		FileFormat:        fileFormats,
		FileNameFormat:    fileNameFormat,
	}, &jobId)
	if err != nil {
		return "", fmt.Errorf("failed to add entity: %w", err)
	}

	JobQueueMessage, err := json.Marshal(dto.PanoramaQueueMessage{
		JobId: jobId,
	})
	if err != nil {
		s.DB.Delete(ctx, jobId)
		return "", fmt.Errorf("failed to serialize queue message: %w", err)
	}

	_, err = s.Queue.Enqueue(ctx, string(JobQueueMessage))
	if err != nil {
		s.DB.Delete(ctx, jobId)
		return "", fmt.Errorf("failed to enqueue message: %w", err)
	}

	return jobId, nil
}

func (s *PanoramaService) GetStatus(ctx context.Context, jobId string) (*dto.PanoramaEntity, error) {
	body, err := s.DB.Find(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	return &body, nil
}

func (s *PanoramaService) GetUploadUrl(ctx context.Context, fileName string, contentType string) (string, string, error) {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return "", "", ErrFileExtensionRequired
	}
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	if !slices.Contains(constants.ValidContentTypeList, contentType) {
		return "", "", ErrInvalidContentType
	}
	if ext != "jpg" && ext != "jpeg" && ext != "png" {
		return "", "", ErrInvalidContentType
	}

	v4, err := uuid.NewRandom()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate v4 uuid: %w", err)
	}

	blobName := v4.String() + "-" + fileName
	duration := time.Hour * 1

	uploadUrl, err := s.PanoramaStorage.GenerateUploadUrl(ctx, blobName, contentType, duration)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate upload url: %w", err)
	}

	return uploadUrl, blobName, nil
}

func (s *PanoramaService) GetShareUrl(ctx context.Context, jobId string) ([]dto.TileDownload, error) {
	body, err := s.DB.Find(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	if body.Status != "Completed" {
		return nil, ErrJobNotComplete
	}

	if body.PanoramaSliceId == nil {
		return nil, ErrJobNotComplete
	}

	var tileDownloads []dto.TileDownload

	for i := range body.Row {
		for j := range body.Column {
			blobName := *body.PanoramaSliceId + "\\" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".jpg"

			var sasUrl string
			sasUrl, err = s.PanoramaSliceStorage.GenerateShareUrl(ctx, blobName, 1*time.Hour)
			if err != nil {
				return nil, err
			}

			row := i + 1
			col := j + 1

			fileNameFormat, err := util.FormatFileName(body.FileNameFormat, row, col)
			if err != nil {
				return nil, err
			}

			tileDownloads = append(tileDownloads, dto.TileDownload{
				UrlSAS:   sasUrl,
				FileName: fileNameFormat + "." + body.FileFormat,
				Row:      row,
				Column:   col,
			})
		}
	}

	return tileDownloads, nil
}

func (s *PanoramaService) GetArchive(ctx context.Context, jobId string) ([]byte, error) {
	body, err := s.DB.Find(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("failed to find job: %w", err)
	}

	if body.Status != "Completed" || body.PanoramaSliceId == nil {
		return nil, ErrJobNotComplete
	}

	var archive bytes.Buffer
	zipWriter := zip.NewWriter(&archive)

	for i := range body.Row {
		for j := range body.Column {
			blobName := *body.PanoramaSliceId + "\\" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".jpg"

			tileName, err := util.FormatFileName(body.FileNameFormat, i+1, j+1)
			if err != nil {
				zipWriter.Close()
				return nil, err
			}

			tileWriter, createErr := zipWriter.Create(tileName + "." + body.FileFormat)
			if createErr != nil {
				zipWriter.Close()
				return nil, createErr
			}

			if downloadErr := s.PanoramaSliceStorage.Download(ctx, blobName, tileWriter); downloadErr != nil {
				zipWriter.Close()
				return nil, downloadErr
			}
		}
	}

	if err = zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return archive.Bytes(), nil
}
