package main

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	err := generateSSHKeyPair()
	if err != nil {
		fmt.Println(err.Error())
	}
	creatGKEComputeEngine()
	time.Sleep(time.Minute * 1)
	fmt.Println("Slept well")
	sshIntoMachine()
}

// Generates an SSH key pair using the 'ssh-keygen' command
func generateSSHKeyPair() error {
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", "id_rsa", "-N", "")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate SSH key pair: %v", err)
	}
	return nil
}

func readPublicKey(publicKeyPath string) string {
	// Read the contents of the SSH public key file
	// Replace with your own file reading implementation
	// Here's an example:
	contents, err := os.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatalf("Failed to read public key: %v", err)
	}
	publicKey := string(contents)

	return publicKey
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
	publicKeyBytes, err := os.ReadFile("id_rsa.pub")
	if err != nil {
		log.Fatalf("Failed to read public key: %v", err)
	}
	publicKey := strings.TrimSpace(string(publicKeyBytes))
	sshKey := fmt.Sprintf("%s:%s", os.Getenv("USER"), publicKey)
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
				{
					Key:   "ssh-keys",
					Value: &sshKey,
				},
			},
		},
	}
}

func sshIntoMachine() {
	host := getInstaceIP()
	port := 22
	user := os.Getenv("USER")
	privateKeyPath := "id_rsa"

	// Read the private key file
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}

	// Parse the private key
	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Configure the SSH client
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// Establish the SSH connection
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		log.Fatalf("Failed to connect to SSH server: %v", err)
	}
	defer conn.Close()

	// Execute a command on the remote machine
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Run a command on the remote machine
	output, err := session.Output("cat /hello.txt")
	if err != nil {
		log.Fatalf("Failed to run command on remote machine: %v", err)
	}

	fmt.Println("Command output:")
	fmt.Println(string(output))
}

func getInstaceIP() string {
	projectID := "appscode-testing"
	// Replace with the name of your instance
	instanceName := "shaad-test"

	ctx := context.Background()

	// Create a compute service client
	gcpCred := os.Getenv("GCP_CREDENTIAL")
	service, err := compute.NewService(ctx, option.WithCredentialsJSON([]byte(gcpCred)))
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}

	// Get the instance details
	instance, err := service.Instances.Get(projectID, "us-central1-a", instanceName).Do()
	if err != nil {
		log.Fatalf("Failed to get instance details: %v", err)
	}

	for _, networkInterface := range instance.NetworkInterfaces {
		for _, accessConfig := range networkInterface.AccessConfigs {
			if accessConfig.NatIP != "" {
				fmt.Println("Instance IP:", accessConfig.NatIP)
				return accessConfig.NatIP
			}
		}
	}

	fmt.Println("No IP address found for the instance.")
	return ""
}

//  export GCP_CREDENTIAL=$(cat /home/shaad/go/src/github.com/Shaad7/capi-basics/gcp/gcp-cred.json  | tr -d '\n')
