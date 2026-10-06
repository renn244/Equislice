package azure

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

func CreateTableInstance(tableName string, connectionString string) (*aztables.Client, error) {
	serviceClient, err := aztables.NewServiceClientFromConnectionString(connectionString, &aztables.ClientOptions{
		ClientOptions: tracingOptions(),
	})
	if err != nil {
		return nil, errors.New("Invallid Connection String")
	}

	client := serviceClient.NewClient(tableName)

	return client, nil
}

func UpdateStatusJobEntity(ctx context.Context, table *aztables.Client, jobId string, status string, panoramaSliceId *string) error {
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

func GetJobEntity(ctx context.Context, table *aztables.Client, jobId string) (*PanoramaEntity, error) {
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
