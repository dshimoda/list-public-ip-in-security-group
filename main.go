package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var privateIpCIDRs = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithDefaultRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	listSecurityGroupIDs(cfg)
}

func listSecurityGroupIDs(cfg aws.Config) {
	cli := ec2.NewFromConfig(cfg)
	params := &ec2.DescribeSecurityGroupsInput{}
	resp, err := cli.DescribeSecurityGroups(context.Background(), params)
	if err != nil {
		log.Fatalf("can not list security group, %v", err)
	}
	for _, sg := range resp.SecurityGroups {
		for _, rule := range sg.IpPermissions {
			formatPrint(rule, *sg.GroupId, *sg.GroupName, "ingress")
		}
		for _, rule := range sg.IpPermissionsEgress {
			formatPrint(rule, *sg.GroupId, *sg.GroupName, "egress")
		}
	}
}

func formatPrint(rule ec2Types.IpPermission, id string, name string, direction string) {
	for _, ip := range rule.IpRanges {
		if ip.Description == nil || !reflect.ValueOf(ip.Description).IsNil() {
			ip.Description = aws.String("empty")
		}
		if rule.FromPort == nil || reflect.ValueOf(rule.FromPort).IsNil() {
			rule.FromPort = aws.Int32(int32(-1))
		}
		if rule.ToPort == nil || reflect.ValueOf(rule.ToPort).IsNil() {
			rule.ToPort = aws.Int32(int32(-1))
		}
		isPrivate := false
		for _, privateCidr := range privateIpCIDRs {
			_, privateRange, _ := net.ParseCIDR(privateCidr)
			ip, _, err := net.ParseCIDR(*ip.CidrIp)
			if err != nil {
				log.Fatalf("unkown cidr subnet, %v", err)
			}
			if privateRange.Contains(ip) {
				isPrivate = true
			}
		}
		if isPrivate {
			continue
		}
		fmt.Printf("%v,%v,%s,%v,%v,%v,%v\n", id, name, direction, *rule.FromPort, *rule.ToPort, *ip.CidrIp, *ip.Description)
	}
}
