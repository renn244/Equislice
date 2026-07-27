package dto

type PanoramaEntity struct {
	JobId             string  `json:"job_id"`
	InitialPanoramaId string  `json:"initial_panorama_id"`
	PanoramaSliceId   *string `json:"panorama_slice_id"`
	Status            string  `json:"status"`
	Row               int     `json:"row"`
	Column            int     `json:"column"`
	FileFormat        string  `json:"file_format"`
}

type PanoramaQueueMessage struct {
	JobId string `json:"job_id"`
}

type PostPanoramaRequest struct {
	File        string `json:"file" binding:"required"`
	Rows        int    `json:"rows" validate:"required,min=1"`
	Columns     int    `json:"columns" validate:"required,min=1"`
	FileFormats string `json:"file_formats" validate:"required"`
}

type PostPanoramaResponse struct {
	JobId string `json:"job_id"`
}

type GetStatusPanoramaResponse struct {
	Status  string `json:"status"`
	Rows    int    `json:"rows"`
	Columns int    `json:"columns"`
}

type GetUploadUrlRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
}

type GetUploadUrlResponse struct {
	UrlSAS   string `json:"urlSAS"`
	BlobName string `json:"blobName"`
}

type GetSASUrlResponse struct {
	UrlsSAS []string `json:"urlsSAS"`
}
