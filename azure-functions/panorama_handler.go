package main

import (
	"azure-function/azure"

	"bytes"
	"context"
	"encoding/json"
	"image"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

func processPanoramaSliceHandler(ctx context.Context, msg bindings.QueueMessage) (handlerErr error) {
	azure.LogQueueItem(ctx, msg)

	var body azure.PanoramaSliceBody

	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error unmarshalling panorama body", "body", string(msg.Body))
		return err
	}

	ctx, endSpan := startPanoramaQueueSpan(ctx, msg, body, "panorama-slice")
	defer func() { endSpan(handlerErr) }()

	connectionString := os.Getenv("EQUISLICE_STORAGE_CONNECTION_STRING")

	// instantiate instances
	panoramaTable, err := azure.CreateTableInstance("Panorama", connectionString)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error creating table instance", "error:", err)
		return err
	}

	blobInstance, err := azure.CreateBlobInstance(connectionString)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error creating blob instance", "error:", err)
		return err
	}

	err = azure.UpdateStatusJobEntity(ctx, panoramaTable, body.JobId, "Processing", nil)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	panoramaEntity, err := azure.GetJobEntity(ctx, panoramaTable, body.JobId)
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
					_, err := azure.GetPropertiesBlob(ctx, blobInstance, "equirectangular-slice", fileName)

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
			err = azure.UpdateStatusJobEntity(ctx, panoramaTable, body.JobId, "Completed", &panoramaEntity.JobId)
			if err != nil {
				sentry.CaptureException(err)

				return err
			}

			return nil
		}
	}

	imageResponse, err := azure.DownloadBlobStream(ctx, blobInstance, "equirectangular", panoramaEntity.InitialPanoramaId)
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

				azure.CleanupExistingBlobs(ctx, blobInstance, existingPanoramas)

				return err
			}

			blobName := panoramaEntity.JobId + "\\" + strconv.Itoa(i) + "_" + strconv.Itoa(j) + ".jpg"
			existingPanoramas = append(existingPanoramas, blobName)

			_, err = azure.UploadBlobStream(ctx, blobInstance, "equirectangular-slice", blobName, sliceBytes)
			if err != nil {
				sentry.CaptureException(err)

				azure.CleanupExistingBlobs(ctx, blobInstance, existingPanoramas)

				return err
			}
		}
	}

	err = azure.UpdateStatusJobEntity(ctx, panoramaTable, body.JobId, "Completed", &panoramaEntity.JobId)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	return nil
}

func processFailedPanoramaSlicedHandler(ctx context.Context, msg bindings.QueueMessage) (handlerErr error) {
	azure.LogQueueItem(ctx, msg)

	var body azure.PanoramaSliceBody

	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error unmarshalling panorama body", "body", string(msg.Body))
		return err
	}

	ctx, endSpan := startPanoramaQueueSpan(ctx, msg, body, "panorama-slice-poison")
	defer func() { endSpan(handlerErr) }()

	connectionString := os.Getenv("EQUISLICE_STORAGE_CONNECTION_STRING")

	panoramaTable, err := azure.CreateTableInstance("Panorama", connectionString)
	if err != nil {
		sentry.CaptureException(err)

		slog.InfoContext(ctx, "Error creating table instance", "error:", err)
		return err
	}

	err = azure.UpdateStatusJobEntity(ctx, panoramaTable, body.JobId, "Failed", &body.JobId)
	if err != nil {
		sentry.CaptureException(err)

		return err
	}

	return nil
}

func startPanoramaQueueSpan(ctx context.Context, msg bindings.QueueMessage, body azure.PanoramaSliceBody, queueName string) (context.Context, func(error)) {
	invocation := trace.SpanContextFromContext(ctx)
	ctx = azure.ExtractQueueContext(ctx, body.TraceContext)
	// Each invocation processes one message, so the producer is its trace parent.
	// Links preserve correlation with both the message and any host invocation span.
	ctx, span := otel.Tracer("azure-function").Start(ctx, queueName+" process",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithLinks(
			trace.Link{SpanContext: trace.SpanContextFromContext(ctx)},
			trace.Link{SpanContext: invocation},
		),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("azure_queue"),
			semconv.MessagingDestinationName(queueName),
			semconv.MessagingOperationTypeProcess,
			semconv.MessagingMessageID(msg.Id),
			attribute.String("panorama.job_id", body.JobId),
			attribute.Int64("azure.queue.dequeue_count", msg.DequeueCount),
		),
	)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "queue processing failed")
		}
		span.End()
	}
}
