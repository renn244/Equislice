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

type PostPanorama struct {
	PanoramaImage string `json:"panorama_image"`
	Rows          int    `json:"rows"`
	Columns       int    `json:"columns"`
	FileFormats   string `json:"file_formats"`
}

type PostPanoramaResponse struct {
	JobId string `json:"job_id"`
}

type GetStatusPanoramaResponse struct {
	Status string `json:"status"`
}

type GetSASUrlResponse struct {
	UrlsSAS []string `json:"urlsSAS"`
}
