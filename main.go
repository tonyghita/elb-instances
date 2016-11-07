package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
)

func main() {
	profile := flag.String("profile", "default", "Set the AWS credential profile to use.")
	region := flag.String("region", "us-west-2", "Set the AWS region to use.")
	max := flag.Int64("n", 1000, "The maximum number of instance IP addresses to return")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("you must specify an ELB name!")
	}

	elbName := args[0]

	sess, err := session.NewSessionWithOptions(session.Options{Profile: *profile})
	if err != nil {
		log.Fatalln("unable to create new session:", err)
	}

	cfg := aws.NewConfig().WithRegion(*region)

	elbsvc := elb.New(sess, cfg)
	res, err := elbsvc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{&elbName},
	})

	if err != nil {
		fmt.Println("profile:", *profile)
		fmt.Println("region:", *region)
		fmt.Println("name:", elbName)

		log.Fatalln("unable to describe load balancer:", err)
	}

	if len(res.LoadBalancerDescriptions) < 1 {
		log.Fatalf("unable to find the load balancer %q\n", elbName)
	}

	instanceIDs := []*string{}
	for _, instance := range res.LoadBalancerDescriptions[0].Instances {
		instanceIDs = append(instanceIDs, instance.InstanceId)
	}

	ec2svc := ec2.New(sess, cfg)
	ec2Res, err := ec2svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	})

	if err != nil {
		log.Fatalln("unable to describe instances:", err)
	}

	var n int64
	for _, r := range ec2Res.Reservations {
		for _, i := range r.Instances {
			fmt.Println(*i.PrivateIpAddress)

			n = n + 1
			if n >= *max {
				os.Exit(0)
			}
		}
	}
}
