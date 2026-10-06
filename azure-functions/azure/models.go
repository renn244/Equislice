package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

type PanoramaSliceBody struct {
	JobId        string            `json:"job_id"`
	TraceContext map[string]string `json:"trace_context,omitempty"`
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
