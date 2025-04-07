package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {

	if len(os.Args) != 3 {
		log.Fatalf("usage: associate-ebs <volume-id> </dev/xvd*>")
	}

	volumeID, deviceName := os.Args[1], os.Args[2]

	if exists(deviceName) {
		log.Printf("Device %q already present, exiting", deviceName)
		return
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("associate-ebs: unable to load config: %v", err)
	}

	imdsClient := imds.NewFromConfig(cfg)

	regionOutput, err := imdsClient.GetRegion(context.Background(), &imds.GetRegionInput{})
	if err != nil {
		log.Fatalf("associate-ebs: unable to determine region failed: %v", err)
	}

	instanceIDOutput, err := imdsClient.GetMetadata(context.Background(), &imds.GetMetadataInput{Path: "instance-id"})
	if err != nil {
		log.Fatalf("associate-ebs: unable to determine instance id: %v", err)
	}
	defer instanceIDOutput.Content.Close()

	instanceIDBytes, err := io.ReadAll(instanceIDOutput.Content)
	if err != nil {
		log.Fatalf("associate-ebs: unable to read instance id bytes: %v", err)
	}
	instanceID := string(instanceIDBytes)

	svc := ec2.NewFromConfig(aws.Config{Region: regionOutput.Region})

	args := &ec2.AttachVolumeInput{
		InstanceId: aws.String(instanceID),
		VolumeId:   aws.String(volumeID),
		Device:     aws.String(deviceName),
	}

	attachment, err := svc.AttachVolume(context.Background(), args)

	if err != nil {
		log.Fatalf("associate-ebs: AttachVolume failed: %v", err)
	}

	log.Printf("Attachment State: %q", attachment.State)

	tick := time.NewTicker(100 * time.Millisecond).C

	timeout := 1 * time.Minute
	deadline := time.After(timeout)
	start := time.Now()

	for {
		select {
		case <-tick:
		case <-deadline:
			log.Fatalf("associate-ebs: device did not attach after %v", timeout)
		}

		if exists(deviceName) {
			log.Printf("Attached in %v", time.Since(start))
			// Success
			return
		}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
