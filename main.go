package main

import (
	"context"
	"fmt"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"log"
	"os"
)

func main() {
	creatGKEComputeEngine()
}

func creatGKEComputeEngine() {
	gcpCred := os.Getenv("GCP_CREDENTIAL")
	clientOption := option.WithCredentialsJSON([]byte(gcpCred))

	computeService, err := compute.NewService(context.Background(), clientOption)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	image, err := computeService.Images.GetFromFamily("ubuntu-os-cloud", "ubuntu-2204-lts").Do()
	if err != nil {
		log.Fatalf("Failed to retrieve Ubuntu 22.04 LTS image: %v", err)
	}

	_, err = computeService.Instances.Insert("appscode-testing", "us-central1-a", getComputeInstance(image.SelfLink)).Context(context.Background()).Do()
	if err != nil {
		log.Fatalf("Failed to create instance: %v", err)
	}
	fmt.Println("Created GKE instance ")
}

func getComputeInstance(image string) *compute.Instance {
	abc, err := os.ReadFile("startup-script.sh")
	if err != nil {
		log.Fatal(err)
	}
	data := string(abc)
	val := "AppsCode"
	return &compute.Instance{
		Name:        "shaad-test",
		MachineType: "projects/appscode-testing/zones/us-central1-a/machineTypes/n1-standard-2",
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    "shaad-disk",
					DiskType:    "zones/us-central1-a/diskTypes/pd-ssd",
					SourceImage: image,
					DiskSizeGb:  10,
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Name: "External NAT",
						Type: "ONE_TO_ONE_NAT",
					},
				},
				Network: "global/networks/default",
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "startup-script",
					Value: &data,
				},
				{
					Key:   "envr",
					Value: &val,
				},
			},
		},
	}

}

//  export GCP_CREDENTIAL=$(cat /home/shaad/go/src/github.com/Shaad7/capi-basics/gcp/gcp-cred.json  | tr -d '\n')
