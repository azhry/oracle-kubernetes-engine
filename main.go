package main

import (
	"echo/helper"
)

var (
	compartmentId      = "ocid1.tenancy.oc1..aaaaaaaaspsylihi2umh6cz3zkkdzkcstcv2kpbjycrqehm34tpudyxufvna"
	clusterId          = "ocid1.cluster.oc1.ap-singapore-1.aaaaaaaalbeyekmulhaf6nbeoutt2e4nm6mruf5fndtqj2bwccq3tf4gj62q"
	vcnDisplayName     = "OCI-GOSDK-Sample-VCN"
	subnetDisplayName1 = "OCI-GOSDK-Sample-Subnet1"
	subnetDisplayName2 = "OCI-GOSDK-Sample-Subnet2"
	subnetDisplayName3 = "OCI-GOSDK-Sample-Subnet3"
)

// https://github.com/oracle/oci-go-sdk/blob/master/example/example_containerengine_test.go
func main() {
	helper.Experiment1()
}
