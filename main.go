package main

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

	s := session.New()

	meta := ec2metadata.New(s)

	region, err := meta.Region()
	if err != nil {
		log.Fatalf("associate-ebs: unable to determine region failed: %v", err)
	}

	instanceID, err := meta.GetMetadata("instance-id")
	if err != nil {
		log.Fatalf("associate-ebs: unable to determine instance id: %v", err)
	}

	svc := ec2.New(s, &aws.Config{Region: aws.String(region)})

	args := &ec2.AttachVolumeInput{
		InstanceId: aws.String(instanceID),
		VolumeId:   aws.String(volumeID),
		Device:     aws.String(deviceName),
	}

	attachment, err := svc.AttachVolume(args)

	if err != nil {
		log.Fatalf("associate-ebs: AttachVolume failed: %v", err)
	}

	log.Printf("Attachment State: %q", *attachment.State)

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
