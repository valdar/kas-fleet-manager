package ocm

import (
	"gitlab.cee.redhat.com/service/managed-services-api/pkg/constants"

	sdkClient "github.com/openshift-online/ocm-sdk-go"
	clustersmgmtv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"gitlab.cee.redhat.com/service/managed-services-api/pkg/api"
	"gitlab.cee.redhat.com/service/managed-services-api/pkg/errors"
)

//go:generate moq -out client_moq.go . Client
type Client interface {
	CreateCluster(cluster *clustersmgmtv1.Cluster) (*clustersmgmtv1.Cluster, error)
	GetClusterIngresses(clusterID string) (*clustersmgmtv1.IngressesListResponse, error)
	GetClusterStatus(id string) (*clustersmgmtv1.ClusterStatus, error)
	GetCloudProviders() (*clustersmgmtv1.CloudProviderList, error)
	GetRegions(provider *clustersmgmtv1.CloudProvider) (*clustersmgmtv1.CloudRegionList, error)
	GetManagedKafkaAddon(id string) (*clustersmgmtv1.AddOnInstallation, error)
	CreateManagedKafkaAddon(id string) (*clustersmgmtv1.AddOnInstallation, error)
	GetClusterDNS(clusterID string) (string, error)
	CreateSyncSet(clusterID string, syncset *clustersmgmtv1.Syncset) (*clustersmgmtv1.Syncset, error)
	DeleteSyncSet(clusterID string, syncsetID string) (int, error)
	ScaleUpComputeNodes(clusterID string) (*clustersmgmtv1.Cluster, error)
	ScaleDownComputeNodes(clusterID string) (*clustersmgmtv1.Cluster, error)
}

var _ Client = &client{}

type client struct {
	ocmClient *sdkClient.Connection
}

func NewClient(ocmClient *sdkClient.Connection) Client {
	return &client{
		ocmClient: ocmClient,
	}
}

func (c *client) CreateCluster(cluster *clustersmgmtv1.Cluster) (*clustersmgmtv1.Cluster, error) {
	clusterResource := c.ocmClient.ClustersMgmt().V1().Clusters()
	response, err := clusterResource.Add().Body(cluster).Send()
	if err != nil {
		return &clustersmgmtv1.Cluster{}, errors.New(errors.ErrorGeneral, err.Error())
	}
	createdCluster := response.Body()

	return createdCluster, nil
}

// GetClusterIngresses sends a GET request to ocm to retrieve the ingresses of an OSD cluster
func (c *client) GetClusterIngresses(clusterID string) (*clustersmgmtv1.IngressesListResponse, error) {
	clusterIngresses := c.ocmClient.ClustersMgmt().V1().Clusters().Cluster(clusterID).Ingresses()
	ingressList, err := clusterIngresses.List().Send()
	if err != nil {
		return nil, err
	}

	return ingressList, nil
}

func (c client) GetClusterStatus(id string) (*clustersmgmtv1.ClusterStatus, error) {
	resp, err := c.ocmClient.ClustersMgmt().V1().Clusters().Cluster(id).Status().Get().Send()
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (c *client) GetCloudProviders() (*clustersmgmtv1.CloudProviderList, error) {
	providersCollection := c.ocmClient.ClustersMgmt().V1().CloudProviders()
	providersResponse, err := providersCollection.List().Send()
	if err != nil {
		return nil, err
	}
	cloudProviderList := providersResponse.Items()
	return cloudProviderList, nil
}

func (c *client) GetRegions(provider *clustersmgmtv1.CloudProvider) (*clustersmgmtv1.CloudRegionList, error) {
	regionsCollection := c.ocmClient.ClustersMgmt().V1().CloudProviders().CloudProvider(provider.ID()).Regions()
	regionsResponse, err := regionsCollection.List().Send()
	if err != nil {
		return nil, err
	}

	regionList := regionsResponse.Items()
	return regionList, nil
}

func (c client) CreateManagedKafkaAddon(id string) (*clustersmgmtv1.AddOnInstallation, error) {
	addon := clustersmgmtv1.NewAddOn().ID(api.ManagedKafkaAddonID)
	addonInstallation, err := clustersmgmtv1.NewAddOnInstallation().Addon(addon).Build()
	if err != nil {
		return nil, err
	}

	resp, err := c.ocmClient.ClustersMgmt().V1().Clusters().Cluster(id).Addons().Add().Body(addonInstallation).Send()
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (c client) GetManagedKafkaAddon(id string) (*clustersmgmtv1.AddOnInstallation, error) {
	resp, err := c.ocmClient.ClustersMgmt().V1().Clusters().Cluster(id).Addons().List().Send()
	if err != nil {
		return nil, err
	}

	managedKafkaAddon := &clustersmgmtv1.AddOnInstallation{}
	resp.Items().Each(func(addOnInstallation *clustersmgmtv1.AddOnInstallation) bool {
		if addOnInstallation.ID() == api.ManagedKafkaAddonID {
			managedKafkaAddon = addOnInstallation
			return false
		}
		return true
	})

	return managedKafkaAddon, nil
}

func (c *client) GetClusterDNS(clusterID string) (string, error) {
	if clusterID == "" {
		return "", errors.New(errors.ErrorGeneral, "ClusterID cannot be empty")
	}
	ingresses, err := c.GetClusterIngresses(clusterID)
	if err != nil {
		return "", errors.New(errors.ErrorGeneral, err.Error())
	}

	var clusterDNS string
	ingresses.Items().Each(func(ingress *clustersmgmtv1.Ingress) bool {
		if ingress.Default() == true {
			clusterDNS = ingress.DNSName()
			return false
		}
		return true
	})
	return clusterDNS, nil
}

func (c client) CreateSyncSet(clusterID string, syncset *clustersmgmtv1.Syncset) (*clustersmgmtv1.Syncset, error) {
	clustersResource := c.ocmClient.ClustersMgmt().V1().Clusters()
	response, syncsetErr := clustersResource.Cluster(clusterID).
		ExternalConfiguration().
		Syncsets().
		Add().
		Body(syncset).
		Send()
	return response.Body(), syncsetErr
}

// Status returns the response status code.
func (c client) DeleteSyncSet(clusterID string, syncsetID string) (int, error) {
	clustersResource := c.ocmClient.ClustersMgmt().V1().Clusters()
	response, syncsetErr := clustersResource.Cluster(clusterID).
		ExternalConfiguration().
		Syncsets().
		Syncset(syncsetID).
		Delete().
		Send()
	return response.Status(), syncsetErr
}

// ScaleUpComputeNodes scales up compute nodes by ClusterNodeScaleIncrement value
func (c client) ScaleUpComputeNodes(clusterID string) (*clustersmgmtv1.Cluster, error) {
	return c.scaleComputeNodes(clusterID, constants.ClusterNodeScaleIncrement)
}

// ScaleDownComputeNodes scales down compute nodes by ClusterNodeScaleIncrement value
func (c client) ScaleDownComputeNodes(clusterID string) (*clustersmgmtv1.Cluster, error) {
	return c.scaleComputeNodes(clusterID, -constants.ClusterNodeScaleIncrement)
}

// scaleComputeNodes scales the Compute nodes up or down by the value of `numNodes`
func (c client) scaleComputeNodes(clusterID string, numNodes int) (*clustersmgmtv1.Cluster, error) {
	clusterClient := c.ocmClient.ClustersMgmt().V1().Clusters().Cluster(clusterID)

	cluster, err := clusterClient.Get().Send()
	if err != nil {
		return nil, err
	}

	// get current number of compute nodes
	currentNumOfNodes := cluster.Body().Nodes().Compute()

	// create a cluster object with updated number of compute nodes
	// NOTE - there is no need to handle whether the number of nodes is valid, as this is handled by OCM
	patch, err := clustersmgmtv1.NewCluster().Nodes(clustersmgmtv1.NewClusterNodes().Compute(currentNumOfNodes + numNodes)).
		Build()
	if err != nil {
		return nil, err
	}

	// patch cluster with updated number of compute nodes
	resp, err := clusterClient.Update().Body(patch).Send()
	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}
