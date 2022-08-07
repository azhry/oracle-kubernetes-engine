## Author
[Azhary Arliansyah](https://github.com/azhry)

## Table of Contents
- [VCN Rough Sketch](#vcn-rough-sketch)
- [Cluster Specs](#cluster-specs)
- [Node Pool Specs](#node-pool-specs)
- [VCN Specs](#vcn-specs)
- [VCN Components](#vcn-components)
- [Steps](#steps)
	* [Create Virtual Cloud Network](#1-create-virtual-cloud-network) 
	* [Create Internet Gateway](#2-create-internet-gateway)
	* [Create NAT Gateway](#3-create-nat-gateway)
	* [Create Service Gateway](#4-create-service-gateway)
	* [Create Private Route Table](#5-create-private-route-table)
	* [Create Public Route Table](#6-create-public-route-table)
	* [Create K8s Security List](#7-create-k8s-security-list)
	* [Create Node Security List](#8-create-node-security-list)
	* [Create Subnets](#9-create-subnets)
	* [Create Cluster](#10-create-cluster)
	* [Migrate to VCN-Native Cluster](#11-migrate-to-vcn-native-cluster)
	* [Create Node Pool](#12-create-node-pool)


## VCN Rough Sketch
![vcn-sketch](https://user-images.githubusercontent.com/9222583/183304079-74e87c7f-09c0-4a57-95d0-83d419e600e0.png)



## Cluster Specs
<ul>
  <li>Kubernetes Version: v1.22.5</li>
  <li>Assign Public IP: true</li>
  <li>Service LB Subnet: svc-subnet</li>
  <li>VCN: OCI-GOSDK-Az-VCN</li>
  <li>Node Count: 3 Nodes</li>
</ul>


## Node Pool Specs
- Node Count: 3
- Kubernetes Version: v1.22.5
- Shape: VM.Standard.E3.Flex
- OCPUs: 4


## VCN Specs
<ul>
  <li>Display Name: OCI-GOSDK-Az-VCN</li>
  <li>CIDR Block: 10.0.0.0/16</li>
  <li>DNS Label: vcndns</li>
</ul>


## VCN Components

| Label | Type  | Specs  |
| ------- | --- | --- |
| [internetgateway](#2-create-internet-gateway) | Internet Gateway | CIDR Block: 0.0.0.0/0 |
| [natgateway](#3-create-nat-gateway) | NAT Gateway | CIDR Block: 0.0.0.0/0 |
| [servicegateway](#4-create-service-gateway) | Service Gateway | - |
| [public-route-table](#6-create-public-route-table) | Routing Table | <ul><li>Network entity: Internet Gateway (0.0.0.0/0)</li></ul> |
| [private-route-table](#5-create-private-route-table) | Routing Table | <ul><li>Network entity: NAT Gateway (0.0.0.0/0)</li><li>Network entity: Service Gateway (all-sin-services-in-oracle-services-network)</li></ul> |
| [svc-subnet](#9-create-subnets) | Subnet | <ul><li>CIDR Block: 10.0.20.0/24</li><li>Routing Table: public-route-table</li><li>DNS Label: svcSubnetDns</li></ul> |
| [k8s-subnet](#9-create-subnets) | Subnet | <ul><li>CIDR Block: 10.0.0.0/28</li><li>Routing Table: public-route-table</li><li>DNS Label: k8sSubnetDns</li><li>Security List: k8s-security-list</li></ul> |
| [node-subnet](#9-create-subnets) | Subnet | <ul><li>CIDR Block: 10.0.10.0/24</li><li>Routing Table: private-route-table</li><li>DNS Label: nodeSubnetDns</li><li>Security List: node-security-list</li></ul> |
| [k8s-security-list](#7-create-k8s-security-list) | Security List | <ul> <li> <div>Ingress Security Rules</div><ol> <li> <ul> <li>Protocol: ICMP</li><li>Source: 10.0.10.0/24</li><li>Description: Path discovery</li><li>Type, Code: 3, 4</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Source: 0.0.0.0/0</li><li>Description: External access to Kubernetes API endpoint</li><li>Destination Port Range: 6443</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Source: 10.0.10.0/24</li><li>Description: Kubernetes worker to Kubernetes API endpoint communication</li><li>Destination Port Range: 6443</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Source: 10.0.10.0/24</li><li>Description: Kubernetes worker to control plane communication</li><li>Destination Port Range: 12250</li></ul> </li></ol> </li><li> <div>Egress Security Rules</div><ol> <li> <ul> <li>Protocol: TCP</li><li>Destination: all-sin-services-in-oracle-services-network</li><li>Destination Type: Service CIDR Block</li><li>Destination Port Range: 443</li><li>Description: Allow Kubernetes Control Plane to communicate with OKE</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Destination: 10.0.10.0/24</li><li>Destination Type: CIDR Block</li><li>Description: All traffic to worker nodes</li></ul> </li><li> <ul> <li>Protocol: ICMP</li><li>Destination: 10.0.10.0/24</li><li>Destination Type: CIDR Block</li><li>Description: Path discovery</li><li>Type, Code: 3, 4</li></ul> </li></ol> </li></ul> |
| [node-security-list](#8-create-node-security-list) | Security List | <ul> <li> <div>Ingress Security Rules</div><ol> <li> <ul> <li>Protocol: ICMP</li><li>Source: 10.0.10.0/24</li><li>Description: Allow pods on one worker node to communicate with pods on other worker nodes</li></ul> </li><li> <ul> <li>Protocol: ICMP</li><li>Source: 10.0.0.0/28</li><li>Description: Path discovery</li><li>Type, Code: 3, 4</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Source: 10.0.0.0/28</li><li>Description: TCP access from Kubernetes Control Plane</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Source: 0.0.0.0/0</li><li>Description: Inbound SSH traffic to worker nodes</li><li>Destination Port Range: 22</li></ul> </li></ol> </li><li> <div>Egress Security Rules</div><ol> <li> <ul> <li>Protocol: "all"</li><li>Destination: 10.0.10.0/24</li><li>Description: Allow pods on one worker node to communicate with pods on other worker nodes</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Destination: 10.0.0.0/28</li><li>Destination Port Range: 6443</li><li>Description: Access to Kubernetes API Endpoint</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Destination: 10.0.0.0/28</li><li>Destination Port Range: 12250</li><li>Description: Kubernetes worker to control plane communication</li></ul> </li><li> <ul> <li>Protocol: ICMP</li><li>Destination: 10.0.0.0/28</li><li>Description: Path discovery</li><li>Type, Code: 3, 4</li></ul> </li><li> <ul> <li>Protocol: TCP</li><li>Destination: all-sin-services-in-oracle-services-network</li><li>Destination Port Range: 443</li><li>Destination Type: Service CIDR Block</li><li>Description: Allow nodes to communicate with OKE to ensure correct start-up and continued functioning</li></ul> </li><li> <ul> <li>Protocol: ICMP</li><li>Destination: 0.0.0.0/0</li><li>Description: ICMP Access from Kubernetes Control Plane</li><li>Type, Code: 3, 4</li></ul> </li><li> <ul> <li>Protocol: "all"</li><li>Destination: 0.0.0.0/0</li><li>Description: Worker Nodes access to Internet</li></ul> </li></ol> </li></ul> |


## Steps

### 1. Create Virtual Cloud Network
```go
func CreateVcn(vcnDisplayName, compartmentId string) core.Vcn {
	log.Println("CREATE VCN")
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	fmt.Println(clerr)
	ctx := context.Background()

	request := core.CreateVcnRequest{}
	request.CidrBlock = common.String("10.0.0.0/16")
	request.CompartmentId = common.String(compartmentId)
	request.DisplayName = common.String(vcnDisplayName)
	request.DnsLabel = common.String("vcndns")

	r, err := c.CreateVcn(ctx, request)
	helpers.FatalIfError(err)
	return r.Vcn
}
```


### 2. Create Internet Gateway
```go
func CreateInternetGateway(vcnId, compartmentId string) core.InternetGateway {
	log.Println("CREATE INTERNET GATEWAY ", vcnId)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	createInternetGatewayRequest := core.CreateInternetGatewayRequest{
		CreateInternetGatewayDetails: core.CreateInternetGatewayDetails{
			CompartmentId: common.String(compartmentId),
			VcnId:         common.String(vcnId),
			IsEnabled:     common.Bool(true),
		},
	}

	igw, err := c.CreateInternetGateway(ctx, createInternetGatewayRequest)
	helpers.FatalIfError(err)

	return igw.InternetGateway
}
```


### 3. Create NAT Gateway
```go
func CreateNatGateway(vcnId, compartmentId string) core.NatGateway {
	log.Println("CREATE NAT GATEWAY ", vcnId)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	createNatGatewayRequest := core.CreateNatGatewayRequest{
		CreateNatGatewayDetails: core.CreateNatGatewayDetails{
			CompartmentId: common.String(compartmentId),
			VcnId:         common.String(vcnId),
		},
	}

	nat, err := c.CreateNatGateway(ctx, createNatGatewayRequest)
	helpers.FatalIfError(err)

	return nat.NatGateway
}
```


### 4. Create Service Gateway
```go
func CreateServiceGateway(vcnId, compartmentId string) core.ServiceGateway {
	log.Println("CREATE SERVICE GATEWAY ", vcnId)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	createServiceGatewayRequest := core.CreateServiceGatewayRequest{
		CreateServiceGatewayDetails: core.CreateServiceGatewayDetails{
			CompartmentId: common.String(compartmentId),
			VcnId:         common.String(vcnId),
			Services:      []core.ServiceIdRequestDetails{},
		},
	}

	sgw, err := c.CreateServiceGateway(ctx, createServiceGatewayRequest)
	helpers.FatalIfError(err)

	return sgw.ServiceGateway
}
```


### 5. Create Private Route Table
```go
func CreatePrivateRouteTable(vcnId, compartmentId, natId, serviceId string) core.RouteTable {
	log.Println("CREATE PRIVATE ROUTE TABLE ", vcnId)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	createRouteTableRequest := core.CreateRouteTableRequest{
		CreateRouteTableDetails: core.CreateRouteTableDetails{
			CompartmentId: common.String(compartmentId),
			VcnId:         common.String(vcnId),
			DisplayName:   common.String("private-route-table"),
			RouteRules: []core.RouteRule{
				{
					NetworkEntityId: common.String(natId), // NAT gateway id
					Description:     common.String("traffic to the internet"),
					Destination:     common.String("0.0.0.0/0"),
					DestinationType: core.RouteRuleDestinationTypeCidrBlock,
				},
				{
					NetworkEntityId: common.String(serviceId), // service gateway id
					Description:     common.String("traffic to OCI services"),
					Destination:     common.String("all-sin-services-in-oracle-services-network"),
					DestinationType: core.RouteRuleDestinationTypeServiceCidrBlock,
				},
			},
		},
	}

	rt, err := c.CreateRouteTable(ctx, createRouteTableRequest)
	helpers.FatalIfError(err)

	return rt.RouteTable
}
```


### 6. Create Public Route Table
```go
func CreatePublicRouteTable(vcnId, compartmentId, igwId string) core.RouteTable {
	log.Println("CREATE PUBLIC ROUTE TABLE ", vcnId)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	createRouteTableRequest := core.CreateRouteTableRequest{
		CreateRouteTableDetails: core.CreateRouteTableDetails{
			CompartmentId: common.String(compartmentId),
			VcnId:         common.String(vcnId),
			DisplayName:   common.String("public-route-table"),
			RouteRules: []core.RouteRule{
				{
					NetworkEntityId: common.String(igwId), // internet gateway id
					Description:     common.String("traffic to/from internet"),
					Destination:     common.String("0.0.0.0/0"),
					DestinationType: core.RouteRuleDestinationTypeCidrBlock,
				},
			},
		},
	}

	rt, err := c.CreateRouteTable(ctx, createRouteTableRequest)
	helpers.FatalIfError(err)

	return rt.RouteTable
}
```


### 7. Create K8s Security List
```go
func CreateK8sSecurityList(vcnId, compartmentId string) core.SecurityList {
	log.Println("CREATE K8S SECURITY LIST ", vcnId)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	createSecurityListRequest := core.CreateSecurityListRequest{
		CreateSecurityListDetails: core.CreateSecurityListDetails{
			CompartmentId: common.String(compartmentId),
			VcnId:         common.String(vcnId),
			DisplayName:   common.String("k8s-security-list"),
			IngressSecurityRules: []core.IngressSecurityRule{
				{
					Protocol:    common.String("1"), // ICMP
					Source:      common.String("10.0.10.0/24"),
					Description: common.String("Path discovery"),
					IcmpOptions: &core.IcmpOptions{
						Type: common.Int(3),
						Code: common.Int(4),
					},
				},
				{
					Protocol:    common.String("6"), // TCP
					Source:      common.String("0.0.0.0/0"),
					Description: common.String("External access to Kubernetes API endpoint"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(6443),
							Min: common.Int(6443),
						},
					},
				},
				{
					Protocol:    common.String("6"), // TCP
					Source:      common.String("10.0.10.0/24"),
					Description: common.String("Kubernetes worker to Kubernetes API endpoint communication"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(6443),
							Min: common.Int(6443),
						},
					},
				},
				{
					Protocol:    common.String("6"), // TCP
					Source:      common.String("10.0.10.0/24"),
					Description: common.String("Kubernetes worker to control plane communication"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(12250),
							Min: common.Int(12250),
						},
					},
				},
			},
			EgressSecurityRules: []core.EgressSecurityRule{
				{
					Protocol:        common.String("6"), // TCP
					Description:     common.String("Allow Kubernetes Control Plane to communicate with OKE"),
					Destination:     common.String("all-sin-services-in-oracle-services-network"),
					DestinationType: core.EgressSecurityRuleDestinationTypeServiceCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(443),
							Min: common.Int(443),
						},
					},
				},
				{
					Protocol:        common.String("6"), // TCP
					Description:     common.String("All traffic to worker nodes"),
					Destination:     common.String("10.0.10.0/24"),
					DestinationType: core.EgressSecurityRuleDestinationTypeCidrBlock,
				},
				{
					Protocol:        common.String("1"), // ICMP
					Description:     common.String("Path discovery"),
					Destination:     common.String("10.0.10.0/24"),
					DestinationType: core.EgressSecurityRuleDestinationTypeCidrBlock,
					IcmpOptions: &core.IcmpOptions{
						Type: common.Int(3),
						Code: common.Int(4),
					},
				},
			},
		},
	}

	s, err := c.CreateSecurityList(ctx, createSecurityListRequest)
	helpers.FatalIfError(err)

	return s.SecurityList
}
```


### 8. Create Node Security List
```go
func CreateNodeSecurityList(vcnId, compartmentId string) core.SecurityList {
	log.Println("CREATE NODE SECURITY LIST ", vcnId)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	createSecurityListRequest := core.CreateSecurityListRequest{
		CreateSecurityListDetails: core.CreateSecurityListDetails{
			CompartmentId: common.String(compartmentId),
			VcnId:         common.String(vcnId),
			DisplayName:   common.String("node-security-list"),
			IngressSecurityRules: []core.IngressSecurityRule{
				{
					Protocol:    common.String("all"), // ICMP
					Source:      common.String("10.0.10.0/24"),
					Description: common.String("Allow pods on one worker node to communicate with pods on other worker nodes"),
				},
				{
					Protocol:    common.String("1"), // ICMP
					Source:      common.String("10.0.0.0/28"),
					Description: common.String("Path discovery"),
					IcmpOptions: &core.IcmpOptions{
						Type: common.Int(3),
						Code: common.Int(4),
					},
				},
				{
					Protocol:    common.String("6"), // TCP
					Source:      common.String("10.0.0.0/28"),
					Description: common.String("TCP access from Kubernetes Control Plane"),
				},
				{
					Protocol:    common.String("6"), // TCP
					Source:      common.String("0.0.0.0/0"),
					Description: common.String("Inbound SSH traffic to worker nodes"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(22),
							Min: common.Int(22),
						},
					},
				},
			},
			EgressSecurityRules: []core.EgressSecurityRule{
				{
					Protocol:    common.String("all"),
					Description: common.String("Allow pods on one worker node to communicate with pods on other worker nodes"),
					Destination: common.String("10.0.10.0/24"),
				},
				{
					Protocol:    common.String("6"), // TCP
					Description: common.String("Access to Kubernetes API Endpoint"),
					Destination: common.String("10.0.0.0/28"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(6443),
							Min: common.Int(6443),
						},
					},
				},
				{
					Protocol:    common.String("6"), // TCP
					Description: common.String("Kubernetes worker to control plane communication"),
					Destination: common.String("10.0.0.0/28"),
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(12250),
							Min: common.Int(12250),
						},
					},
				},
				{
					Protocol:    common.String("1"), // ICMP
					Description: common.String("Path discovery"),
					Destination: common.String("10.0.0.0/28"),
					IcmpOptions: &core.IcmpOptions{
						Type: common.Int(3),
						Code: common.Int(4),
					},
				},
				{
					Protocol:        common.String("6"), // TCP
					Description:     common.String("Allow nodes to communicate with OKE to ensure correct start-up and continued functioning"),
					Destination:     common.String("all-sin-services-in-oracle-services-network"),
					DestinationType: core.EgressSecurityRuleDestinationTypeServiceCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(443),
							Min: common.Int(443),
						},
					},
				},
				{
					Protocol:    common.String("1"), // ICMP
					Description: common.String("ICMP Access from Kubernetes Control Plane"),
					Destination: common.String("0.0.0.0/0"),
					IcmpOptions: &core.IcmpOptions{
						Type: common.Int(3),
						Code: common.Int(4),
					},
				},
				{
					Protocol:    common.String("all"),
					Description: common.String("Worker Nodes access to Internet"),
					Destination: common.String("0.0.0.0/0"),
				},
			},
		},
	}

	s, err := c.CreateSecurityList(ctx, createSecurityListRequest)
	helpers.FatalIfError(err)

	return s.SecurityList
}
```


### 9. Create Subnets
```go
func CreateSubnet(displayName *string, cidrBlock *string, dnsLabel *string, availableDomain *string, vcn core.Vcn, routeTable core.RouteTable, securityList *core.SecurityList) core.Subnet {
	log.Println("CREATE SUBNET ", *displayName)
	c, clerr := core.NewVirtualNetworkClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	ctx := context.Background()

	request := core.CreateSubnetRequest{}
	request.AvailabilityDomain = availableDomain
	request.CompartmentId = common.String(compartmentId)
	request.CidrBlock = cidrBlock
	request.DisplayName = displayName
	request.DnsLabel = dnsLabel
	request.RequestMetadata = helpers.GetRequestMetadataWithDefaultRetryPolicy()
	request.VcnId = vcn.Id
	request.RouteTableId = routeTable.Id
	if securityList != nil {
		log.Println("SUBNET SECURITY LIST ID:", *securityList.Id)
		request.SecurityListIds = []string{*securityList.Id}
	}

	r, err := c.CreateSubnet(ctx, request)
	helpers.FatalIfError(err)
	log.Println("Subnet created")

	// retry condition check, stop until return true
	pollUntilAvailable := func(r common.OCIOperationResponse) bool {
		if converted, ok := r.Response.(core.GetSubnetResponse); ok {
			log.Println(converted.LifecycleState)
			return converted.LifecycleState != core.SubnetLifecycleStateAvailable
		}
		return true
	}

	pollGetRequest := core.GetSubnetRequest{
		SubnetId:        r.Id,
		RequestMetadata: helpers.GetRequestMetadataWithCustomizedRetryPolicy(pollUntilAvailable),
	}

	// wait for lifecyle become running
	_, pollErr := c.GetSubnet(ctx, pollGetRequest)
	helpers.FatalIfError(pollErr)

	return r.Subnet
}

svcSubnet := CreateSubnet(&subnetDisplayName3, common.String("10.0.20.0/24"), common.String("svcSubnetDns"), nil, vcn, publicRouteTable, nil) // svc subnet

k8sSecurityList := CreateK8sSecurityList(*vcn.Id, compartmentId)
k8sSubnet := CreateSubnet(&subnetDisplayName1, common.String("10.0.0.0/28"), common.String("k8sSubnetDns"), nil, vcn, publicRouteTable, &k8sSecurityList)

nodeSecurityList := CreateNodeSecurityList(*vcn.Id, compartmentId)
_ = CreateSubnet(&subnetDisplayName2, common.String("10.0.10.0/24"), common.String("nodeSubnetDns"), nil, vcn, privateRouteTable, &nodeSecurityList)
```


### 10. Create Cluster
```go
func CreateCluster(vcnId, compartmentId, svcSubnetId, displayName, kubernetesVersion string) containerengine.CreateClusterResponse {
	log.Println("CREATE CLUSTER")

	ctx := context.Background()
	c, clerr := containerengine.NewContainerEngineClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	createClusterRequest := containerengine.CreateClusterRequest{
		CreateClusterDetails: containerengine.CreateClusterDetails{
			Name:              common.String(displayName),
			CompartmentId:     common.String(compartmentId),
			VcnId:             common.String(vcnId),
			KubernetesVersion: common.String(kubernetesVersion),
			Options: &containerengine.ClusterCreateOptions{
				ServiceLbSubnetIds: []string{svcSubnetId},
			},
		},
	}
	resp, err := c.CreateCluster(ctx, createClusterRequest)
	helpers.FatalIfError(err)

	return resp
}
```


### 11. Migrate to VCN-Native Cluster
```go
func MigrateToVcnNativeCluster(clusterId, k8sSubnetId string) containerengine.ClusterMigrateToNativeVcnResponse {
	log.Println("MIGRATE CLUSTER TO NATIVE-VCN")
	ctx := context.Background()
	c, clerr := containerengine.NewContainerEngineClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)
	migrateRequest := containerengine.ClusterMigrateToNativeVcnRequest{
		ClusterId: &clusterId,
		ClusterMigrateToNativeVcnDetails: containerengine.ClusterMigrateToNativeVcnDetails{
			EndpointConfig: &containerengine.ClusterEndpointConfig{
				SubnetId:          &k8sSubnetId,
				IsPublicIpEnabled: common.Bool(true),
			},
		},
	}
	resp, err := c.ClusterMigrateToNativeVcn(ctx, migrateRequest)
	helpers.FatalIfError(err)

	return resp
}
```


### 12. Create Node Pool
```go
func CreateNodePool(nodePoolName, kubernetesVersion, clusterId, imageId, compartmentId string, subnet core.Subnet, ads identity.ListAvailabilityDomainsResponse) {
	log.Println("CREATE NODE POOL ", nodePoolName)
	size := len(ads.Items)
	createNodePoolRequest := containerengine.CreateNodePoolRequest{
		CreateNodePoolDetails: containerengine.CreateNodePoolDetails{
			CompartmentId:     &compartmentId,
			ClusterId:         &clusterId,
			Name:              &nodePoolName,
			KubernetesVersion: &kubernetesVersion,
			NodeShape:         common.String("VM.Standard.E3.Flex"),
			NodeShapeConfig: &containerengine.CreateNodeShapeConfigDetails{
				Ocpus: common.Float32(4),
			},
			NodeConfigDetails: &containerengine.CreateNodePoolNodeConfigDetails{
				Size:             &size,
				PlacementConfigs: make([]containerengine.NodePoolPlacementConfigDetails, 0, size),
			},
			InitialNodeLabels: []containerengine.KeyValue{{Key: common.String("name"), Value: common.String(nodePoolName)}},
			NodeSourceDetails: containerengine.NodeSourceViaImageDetails{ImageId: common.String(imageId)},
		},
	}

	for i := 0; i < len(ads.Items); i++ {
		createNodePoolRequest.NodeConfigDetails.PlacementConfigs = append(createNodePoolRequest.NodeConfigDetails.PlacementConfigs, containerengine.NodePoolPlacementConfigDetails{
			AvailabilityDomain: ads.Items[i].Name,
			SubnetId:           subnet.Id,
		})
	}

	c, clerr := containerengine.NewContainerEngineClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(clerr)

	ctx := context.Background()
	createNodePoolResponse, err := c.CreateNodePool(ctx, createNodePoolRequest)
	helpers.FatalIfError(err)
	fmt.Println("creating nodepool")

	waitUntilWorkRequestComplete(c, createNodePoolResponse.OpcWorkRequestId)
	fmt.Println("nodepool created")
}
```




