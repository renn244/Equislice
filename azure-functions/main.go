package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/getsentry/sentry-go"

	"github.com/azure/azure-functions-golang-worker/sdk"
	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
	"github.com/azure/azure-functions-golang-worker/worker"
)

func main() {
	app := sdk.FunctionApp()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Environment: os.Getenv("SENTRY_ENVIRONMENT"),
	})
	if err != nil {
		panic(err)
	}
	defer sentry.Flush(2 * time.Second)

	app.Queue("processPanoramaSlice", processPanoramaSliceHandler,
		sdk.WithQueueName("panorama-slice"),
		sdk.WithConnection("AzureWebJobsStorage"),
	)

	app.Queue("processFailedPanoramaSliced", processFailedPanoramaSlicedHandler,
		sdk.WithQueueName("panorama-slice-poison"),
		sdk.WithConnection("AzureWebJobsStorage"),
	)

	worker.Start(app)
}

type PanoramaSliceBody struct {
	JobId string `json:"job_id"`
}

type PanoramaEntity struct {
	aztables.Entity

	JobId             string  `json:"job_id"`
	InitialPanoramaId string  `json:"initial_panorama_id"`
	PanoramaSliceId   *string `json:"panorama_slice_id"`
	Status            string  `json:"status"`
	Row               int     `json:"row"`
	Column            int     `json:"column"`
	FileFormat        string  `json:"file_format"`
}

func processPanoramaSliceHandler(ctx context.Context, msg bindings.QueueMessage) error {
	logQueueItem(ctx, msg)

	var body PanoramaSliceBody

	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error unmarshalling panorama body", "body", string(msg.Body))
		return err
	}

	connectionString := os.Getenv("EQUISLICE_STORAGE_CONNECTION_STRING")

	// instantiate instances
	panoramaTable, err := createTableInstance("Panorama", connectionString)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error creating table instance", "error:", err)
		return err
	}

	blobInstance, err := createBlobInstance(connectionString)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error creating blob instance", "error:", err)
		return err
	}

	err = updateStatusJobEntity(ctx, panoramaTable, body.JobId, "Processing", nil)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	panoramaEntity, err := getJobEntity(ctx, panoramaTable, body.JobId)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	if msg.DequeueCount > 1 {

		var tile_fileNames []string
		var isABlobMissing bool = false
		var blobError error

		for i := range panoramaEntity.Row {
			for j := range panoramaEntity.Column {
				blobName := panoramaEntity.JobId + "\\" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".jpg"

				tile_fileNames = append(tile_fileNames, blobName)
			}
		}

		type blobCheckResult struct {
			exists bool
			err    error
		}

		jobs := make(chan string)
		results := make(chan blobCheckResult)

		var wg sync.WaitGroup
		wg.Add(6)

		for range 6 {
			go func() {
				defer wg.Done()

				for fileName := range jobs {
					_, err := getPropertiesBlob(ctx, blobInstance, "equirectangular-slice", fileName)

					if err != nil {
						if bloberror.HasCode(err, bloberror.BlobNotFound) {
							results <- blobCheckResult{exists: false}
							continue
						}

						results <- blobCheckResult{err: err}
						continue
					}

					results <- blobCheckResult{exists: true}
				}
			}()
		}

		go func() {
			for _, fileName := range tile_fileNames {
				jobs <- fileName
			}

			close(jobs)
		}()

		go func() {
			wg.Wait()

			close(results)
		}()

		for result := range results {
			if result.err != nil && blobError == nil {
				blobError = result.err
			}

			if !result.exists {
				isABlobMissing = true
			}
		}

		if blobError != nil {
			return blobError
		}

		if !isABlobMissing {
			err = updateStatusJobEntity(ctx, panoramaTable, body.JobId, "Completed", &panoramaEntity.JobId)
			if err != nil {
				sentry.CaptureException(err)

				return err
			}

			return nil
		}
	}

	imageResponse, err := downloadBlobStream(ctx, blobInstance, "equirectangular", panoramaEntity.InitialPanoramaId)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	panoramaImage := bytes.NewReader(imageResponse)

	config, _, err := image.DecodeConfig(panoramaImage)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}
	img, _, err := image.Decode(bytes.NewReader(imageResponse))
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	heightDimension := config.Height / panoramaEntity.Row
	widthDimension := config.Width / panoramaEntity.Column

	// LATER: check for remainder and put it on the last as an easy solution
	var existingPanoramas []string

	for i := range panoramaEntity.Row {

		y0 := i * heightDimension
		y1 := (i + 1) * heightDimension

		for j := range panoramaEntity.Column {
			x0 := j * widthDimension
			x1 := (j + 1) * widthDimension

			sliceBytes, err := cropSlice(x0, y0, x1, y1, img)
			if err != nil {
				sentry.CaptureException(err)

				cleanupExistingBlobs(ctx, blobInstance, existingPanoramas)

				return err
			}

			blobName := panoramaEntity.JobId + "\\" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".jpg"
			existingPanoramas = append(existingPanoramas, blobName)

			_, err = uploadBlobStream(ctx, blobInstance, "equirectangular-slice", blobName, sliceBytes)
			if err != nil {
				sentry.CaptureException(err)

				cleanupExistingBlobs(ctx, blobInstance, existingPanoramas)

				return err
			}
		}
	}

	err = updateStatusJobEntity(ctx, panoramaTable, body.JobId, "Completed", &panoramaEntity.JobId)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	return nil
}

func processFailedPanoramaSlicedHandler(ctx context.Context, msg bindings.QueueMessage) error {
	logQueueItem(ctx, msg)

	var body PanoramaSliceBody

	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error unmarshalling panorama body", "body", string(msg.Body))
		return err
	}

	connectionString := os.Getenv("EQUISLICE_STORAGE_CONNECTION_STRING")

	panoramaTable, err := createTableInstance("Panorama", connectionString)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error creating table instance", "error:", err)
		return err
	}

	err = updateStatusJobEntity(ctx, panoramaTable, body.JobId, "Failed", &body.JobId)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	return nil
}

func logQueueItem(ctx context.Context, msg bindings.QueueMessage) {
	slog.InfoContext(ctx, "Queue item found!",
		"id", msg.Id,
		"body", string(msg.Body),
		"dequeue_count", msg.DequeueCount,
		"pop_receipt", msg.PopReceipt,
		"expiration_time", msg.ExpirationTime,
		"insertion_time", msg.InsertionTime,
		"next_visible_time", msg.NextVisibleTime,
	)
}

func cleanupExistingBlobs(ctx context.Context, blob *azblob.Client, existingBlobs []string) {
	for _, blobName := range existingBlobs {
		err := deleteBlob(ctx, blob, "equirectangular-slice", blobName)
		if err != nil {
			sentry.CaptureException(err)
		}
	}

}

func cropSlice(x0 int, y0 int, x1 int, y1 int, img image.Image) ([]byte, error) {
	croppedDimension := image.Rect(x0, y0, x1, y1)
	tile := image.NewRGBA(image.Rect(0, 0, croppedDimension.Dx(), croppedDimension.Dy()))

	draw.Draw(tile, tile.Bounds(), img, croppedDimension.Min, draw.Src)

	var buf bytes.Buffer

	err := jpeg.Encode(&buf, tile, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}
	buffers := buf.Bytes()

	return buffers, nil
}

func createTableInstance(tableName string, connectionString string) (*aztables.Client, error) {
	serviceClient, err := aztables.NewServiceClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, errors.New("Invallid Connection String")
	}

	client := serviceClient.NewClient(tableName)

	return client, nil
}

func updateStatusJobEntity(ctx context.Context, table *aztables.Client, jobId string, status string, panoramaSliceId *string) error {
	jobEntity := aztables.EDMEntity{
		Entity: aztables.Entity{
			PartitionKey: "Panoramas",
			RowKey:       jobId,
		},
		Properties: map[string]interface{}{
			"status":            status,
			"panorama_slice_id": panoramaSliceId,
		},
	}
	marshalledEntity, err := json.Marshal(jobEntity)
	if err != nil {
		return err
	}

	_, err = table.UpdateEntity(ctx, marshalledEntity, nil)
	if err != nil {
		return err
	}

	return nil
}

func getJobEntity(ctx context.Context, table *aztables.Client, jobId string) (*PanoramaEntity, error) {
	getEntityResponse, err := table.GetEntity(ctx, "Panoramas", jobId, nil)
	if err != nil {
		// handle does not exist error and some stuff
		// nd handle general error

		return nil, err
	}

	var panoramaEntity PanoramaEntity

	err = json.Unmarshal(getEntityResponse.Value, &panoramaEntity)
	if err != nil {
		return nil, err
	}

	return &panoramaEntity, nil
}

func createBlobInstance(connectionString string) (*azblob.Client, error) {
	if connectionString == "" {
		return nil, errors.New("Invalid Connection String")
	}

	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, errors.New("Error Connecting to Blob")
	}

	return client, nil
}

func downloadBlobStream(ctx context.Context, blob *azblob.Client, containerName string, blobName string) ([]byte, error) {
	response, err := blob.DownloadStream(ctx, containerName, blobName, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	imageResponse, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return imageResponse, nil
}

func getPropertiesBlob(ctx context.Context, blob *azblob.Client, containerName string, blobName string) (*blob.GetPropertiesResponse, error) {
	containerClient := blob.ServiceClient().NewContainerClient(containerName)
	tileClient := containerClient.NewBlockBlobClient(blobName)
	data, err := tileClient.GetProperties(ctx, nil)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func uploadBlobStream(ctx context.Context, blob *azblob.Client, containerName string, blobName string, blobData []byte) (*azblob.UploadStreamResponse, error) {
	reader := bytes.NewReader(blobData)
	response, err := blob.UploadStream(ctx, containerName, blobName, reader, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func deleteBlob(ctx context.Context, blob *azblob.Client, containerName string, blobName string) error {
	_, err := blob.DeleteBlob(ctx, containerName, blobName, nil)
	if err != nil {
		return err
	}

	return nil
}
