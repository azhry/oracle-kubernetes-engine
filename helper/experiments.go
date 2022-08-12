package helper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

var (
	compartmentId      = "ocid1.tenancy.oc1..aaaaaaaaspsylihi2umh6cz3zkkdzkcstcv2kpbjycrqehm34tpudyxufvna"
	clusterName        = "clustersdk"
	vcnDisplayName     = "clustersdkvcn"
	subnetDisplayName1 = "OCI-GOSDK-Az-k8sSubnet"
	subnetDisplayName2 = "OCI-GOSDK-Az-nodeSubnet"
	subnetDisplayName3 = "OCI-GOSDK-Az-svcSubnet"
	nodePoolName       = "pool1"
	kubeVersion        = "v1.22.5"
)

func Experiment1() {
	ctx := context.Background()
	c, clerr := containerengine.NewContainerEngineClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)

	compute, err := core.NewComputeClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)
	req := identity.ListAvailabilityDomainsRequest{}
	req.CompartmentId = common.String(compartmentId)
	ads, err := identityClient.ListAvailabilityDomains(ctx, req)
	helpers.FatalIfError(err)

	vcn := CreateVcn(vcnDisplayName, compartmentId, clusterName)

	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	internetGateway := CreateInternetGateway(*vcn.Id, compartmentId)
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	natGateway := CreateNatGateway(*vcn.Id, compartmentId)
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	serviceGateway := CreateServiceGateway(*vcn.Id, compartmentId)
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	publicRouteTable := CreatePublicRouteTable(*vcn.Id, compartmentId, *internetGateway.Id)
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	privateRouteTable := CreatePrivateRouteTable(*vcn.Id, compartmentId, *natGateway.Id, *serviceGateway.Id)
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	svcSubnet := CreateSubnet(&subnetDisplayName3, common.String("10.0.20.0/24"), common.String("svcSubnetDns"), nil, vcn, publicRouteTable, nil) // svc subnet
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	k8sSecurityList := CreateK8sSecurityList(*vcn.Id, compartmentId)
	k8sSubnet := CreateSubnet(&subnetDisplayName1, common.String("10.0.0.0/28"), common.String("k8sSubnetDns"), nil, vcn, publicRouteTable, &k8sSecurityList)
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	nodeSecurityList := CreateNodeSecurityList(*vcn.Id, compartmentId)
	nodeSubnet := CreateSubnet(&subnetDisplayName2, common.String("10.0.10.0/24"), common.String("nodeSubnetDns"), nil, vcn, privateRouteTable, &nodeSecurityList)
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	createClusterResponse := CreateCluster(ctx, c, *vcn.Id, compartmentId, *svcSubnet.Id, *k8sSubnet.Id, clusterName, kubeVersion)

	// wait until work request complete
	workReqResp := waitUntilWorkRequestComplete(c, createClusterResponse.OpcWorkRequestId)
	fmt.Println("cluster created")
	fmt.Println("WAITING")
	time.Sleep(5 * 60 * time.Second)

	clusterID := getResourceID(workReqResp.Resources, containerengine.WorkRequestResourceActionTypeCreated, "CLUSTER")

	// id := k8sSubnet.Id
	// fmt.Println("ID:", *clusterID, *id)
	// migrateClusterResponse := MigrateToVcnNativeCluster(ctx, c, *clusterID, *id)
	// migreateReqResp := waitUntilWorkRequestComplete(c, migrateClusterResponse.OpcWorkRequestId)
	// fmt.Println("cluster migrated")

	// // wait until migrate complete
	// getResourceID(migreateReqResp.Resources, containerengine.WorkRequestResourceActionTypeCreated, "CLUSTER")

	// // get Image Id
	image := getImageID(ctx, compute)

	fmt.Println(image)
	CreateNodePool(nodePoolName, kubeVersion, *clusterID, *image.Id, compartmentId, *nodeSubnet.Id, ads)

	// AFTER COMPLETION: create tutorial on GitHub and Notion
}

func waitUntilWorkRequestComplete(client containerengine.ContainerEngineClient, workReuqestID *string) containerengine.GetWorkRequestResponse {
	// retry GetWorkRequest call until TimeFinished is set
	shouldRetryFunc := func(r common.OCIOperationResponse) bool {
		return r.Response.(containerengine.GetWorkRequestResponse).TimeFinished == nil
	}

	getWorkReq := containerengine.GetWorkRequestRequest{
		WorkRequestId:   workReuqestID,
		RequestMetadata: helpers.GetRequestMetadataWithCustomizedRetryPolicy(shouldRetryFunc),
	}

	getResp, err := client.GetWorkRequest(context.Background(), getWorkReq)
	helpers.FatalIfError(err)
	return getResp
}

// getResourceID return a resource ID based on the filter of resource actionType and entityType
func getResourceID(resources []containerengine.WorkRequestResource, actionType containerengine.WorkRequestResourceActionTypeEnum, entityType string) *string {
	for _, resource := range resources {
		if resource.ActionType == actionType && strings.ToUpper(*resource.EntityType) == entityType {
			return resource.Identifier
		}
	}

	fmt.Println("cannot find matched resources")
	return nil
}

func getImageID(ctx context.Context, c core.ComputeClient) core.Image {
	request := core.ListImagesRequest{
		CompartmentId: common.String(compartmentId),
		// OperatingSystem: common.String("Oracle Linux"),
		// Shape:           common.String("VM.Standard.E3.Flex"),
		DisplayName: common.String("Oracle-Linux-7.9-2022.06.30-0"),
	}

	r, err := c.ListImages(ctx, request)
	helpers.FatalIfError(err)

	return r.Items[0]
}
