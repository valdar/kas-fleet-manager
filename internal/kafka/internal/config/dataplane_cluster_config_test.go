package config

import (
	"testing"

	"gopkg.in/yaml.v2"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/cloudproviders"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/kafkas/types"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"

	"github.com/onsi/gomega"
)

func TestDataplaneClusterConfig_IsDataPlaneAutoScalingEnabled(t *testing.T) {
	type fields struct {
		DataPlaneClusterScalingType string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "auto scaling enabled",
			fields: fields{
				DataPlaneClusterScalingType: AutoScaling,
			},
			want: true,
		},
		{
			name: "auto scaling disabled",
			fields: fields{
				DataPlaneClusterScalingType: ManualScaling,
			},
			want: false,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := DataplaneClusterConfig{
				DataPlaneClusterScalingType: tt.fields.DataPlaneClusterScalingType,
			}
			g.Expect(conf.IsDataPlaneAutoScalingEnabled()).To(gomega.Equal(tt.want))
		})
	}
}

func TestDataplaneClusterConfig_IsDataPlaneManualScalingEnabled(t *testing.T) {
	type fields struct {
		DataPlaneClusterScalingType string
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "manual scaling enabled",
			fields: fields{
				DataPlaneClusterScalingType: ManualScaling,
			},
			want: true,
		},
		{
			name: "manual scaling disabled",
			fields: fields{
				DataPlaneClusterScalingType: AutoScaling,
			},
			want: false,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := DataplaneClusterConfig{
				DataPlaneClusterScalingType: tt.fields.DataPlaneClusterScalingType,
			}
			got := conf.IsDataPlaneManualScalingEnabled()
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestDataplaneClusterConfig_IsWithinClusterLimit(t *testing.T) {
	type fields struct {
		DataPlaneClusterScalingType string
		ClusterList                 ClusterList
	}

	type args struct {
		clusterId string
		count     int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "within the limit",
			fields: fields{
				DataPlaneClusterScalingType: ManualScaling,
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", KafkaInstanceLimit: 3},
				},
			},
			args: args{
				clusterId: "test01",
				count:     3,
			},
			want: true,
		},
		{
			name: "within the limit if clusterId not in the list",
			fields: fields{
				DataPlaneClusterScalingType: ManualScaling,
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", KafkaInstanceLimit: 3},
				},
			},
			args: args{
				clusterId: "test02",
				count:     3,
			},
			want: true,
		},
		{
			name: "exceed the limit",
			fields: fields{
				DataPlaneClusterScalingType: ManualScaling,
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", KafkaInstanceLimit: 3},
				},
			},
			args: args{
				clusterId: "test01",
				count:     4,
			},
			want: false,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := NewClusterConfig(tt.fields.ClusterList)
			g.Expect(conf.IsNumberOfKafkaWithinClusterLimit(tt.args.clusterId, tt.args.count)).To(gomega.Equal(tt.want))
		})
	}
}

func TestDataplaneClusterConfig_IsClusterSchedulable(t *testing.T) {
	type fields struct {
		ClusterList ClusterList
	}
	type args struct {
		clusterId string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "schedulable",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: true},
				},
			},
			args: args{
				clusterId: "test01",
			},
			want: true,
		},
		{
			name: "schedulable if clusterId not in the config",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: true},
				},
			},
			args: args{
				clusterId: "test02",
			},
			want: true,
		},
		{
			name: "unschedulable",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: false},
				},
			},
			args: args{
				clusterId: "test01",
			},
			want: false,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := NewClusterConfig(tt.fields.ClusterList)
			g.Expect(conf.IsClusterSchedulable(tt.args.clusterId)).To(gomega.Equal(tt.want))
		})
	}
}

func TestDataplaneClusterConfig_MissingClusters(t *testing.T) {
	type fields struct {
		ClusterList ClusterList
	}
	type args struct {
		clusterList map[string]api.Cluster
	}
	var emptyResult []ManualCluster
	var result []ManualCluster
	result = append(result, ManualCluster{ClusterId: "test02", Region: "us-east", MultiAZ: true, CloudProvider: "aws"})
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []ManualCluster
	}{
		{
			name: "Missing clusters found",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test02", Region: "us-east", MultiAZ: true, CloudProvider: "aws"},
				},
			},
			args: args{
				clusterList: map[string]api.Cluster{
					"test01": {
						ClusterID: "test01",
					},
				},
			},
			want: result,
		},
		{
			name: "No Missing clusters found",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test02", Region: "us-east", MultiAZ: true, CloudProvider: "aws"},
				},
			},
			args: args{
				clusterList: map[string]api.Cluster{
					"test02": {
						ClusterID: "test02",
					},
				},
			},
			want: emptyResult,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := NewClusterConfig(tt.fields.ClusterList)
			g.Expect(conf.MissingClusters(tt.args.clusterList)).To(gomega.Equal(tt.want))
		})
	}
}

func TestDataplaneClusterConfig_ExcessClusters(t *testing.T) {
	type fields struct {
		ClusterList ClusterList
	}
	type args struct {
		clusterList map[string]api.Cluster
	}
	var emptyResult []string
	var result []string
	result = append(result, "test01")
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Excess clusters find",
			fields: fields{
				ClusterList{
					ManualCluster{ClusterId: "test02"},
				},
			},
			args: args{
				clusterList: map[string]api.Cluster{
					"test01": {ClusterID: "test01"},
				},
			},
			want: result,
		},
		{
			name: "No Excess clusters find",
			fields: fields{
				ClusterList{
					ManualCluster{ClusterId: "test01"},
				},
			},
			args: args{
				clusterList: map[string]api.Cluster{
					"test01": {ClusterID: "test01"},
				},
			},
			want: emptyResult,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := NewClusterConfig(tt.fields.ClusterList)
			g.Expect(conf.ExcessClusters(tt.args.clusterList)).To(gomega.Equal(tt.want))
		})
	}
}

func TestClusterConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  ManualCluster
		wantErr bool
	}{
		{
			name: "should have default value set",
			input: `
---
name: "test"
cluster_id: "test"
cloud_provider: "aws"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
`,
			output: ManualCluster{
				Name:                  "test",
				ClusterId:             "test",
				CloudProvider:         "aws",
				Region:                "east-1",
				MultiAZ:               true,
				Schedulable:           true,
				KafkaInstanceLimit:    1,
				Status:                api.ClusterProvisioning,
				ProviderType:          api.ClusterProviderOCM,
				SupportedInstanceType: api.AllInstanceTypeSupport.String(),
			},
			wantErr: false,
		},
		{
			name: "should return an error if no cluster_id is set for standalone cluster",
			input: `
---
name: "test"
cloud_provider: "aws"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
`,
			output: ManualCluster{
				Name:                  "test",
				CloudProvider:         "aws",
				Region:                "east-1",
				MultiAZ:               true,
				Schedulable:           true,
				KafkaInstanceLimit:    1,
				Status:                api.ClusterProvisioning,
				ProviderType:          api.ClusterProviderOCM,
				SupportedInstanceType: api.AllInstanceTypeSupport.String(),
			},
			wantErr: true,
		},
		{
			name: "should return an error if no cluster_dns is set for standalone cluster",
			input: `
---
name: "test"
cluster_id: "test"
cloud_provider: "aws"
region: "east-1"
multi_az: true
schedulable: true
provider_type: "standalone"
kafka_instance_limit: 1
`,
			output: ManualCluster{
				Name:                  "test",
				ClusterId:             "test",
				CloudProvider:         "aws",
				Region:                "east-1",
				MultiAZ:               true,
				Schedulable:           true,
				KafkaInstanceLimit:    1,
				Status:                api.ClusterProvisioning,
				ProviderType:          api.ClusterProviderStandalone,
				SupportedInstanceType: api.AllInstanceTypeSupport.String(),
			},
			wantErr: true,
		},
		{
			name: "should return no error if ProviderType is standalone. Status provisioning should be enforced",
			input: `
---
name: "test"
cluster_id: "test"
cloud_provider: "aws"
cluster_dns: "test"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
provider_type: "standalone"
supported_instance_type: "developer"
`,
			output: ManualCluster{
				Name:                  "test",
				ClusterId:             "test",
				CloudProvider:         "aws",
				ClusterDNS:            "test",
				Region:                "east-1",
				MultiAZ:               true,
				Schedulable:           true,
				KafkaInstanceLimit:    1,
				Status:                api.ClusterProvisioning,
				ProviderType:          api.ClusterProviderStandalone,
				SupportedInstanceType: api.DeveloperTypeSupport.String(),
			},
			wantErr: false,
		},
		{
			name: "should assign all instance types if supported_instance_type value is empty",
			input: `
---
name: "test"
cluster_id: "test"
cloud_provider: "aws"
cluster_dns: "test"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
supported_instance_type: ""
provider_type: "standalone"
`,
			output: ManualCluster{
				Name:                  "test",
				ClusterId:             "test",
				CloudProvider:         "aws",
				ClusterDNS:            "test",
				Region:                "east-1",
				MultiAZ:               true,
				Schedulable:           true,
				KafkaInstanceLimit:    1,
				Status:                api.ClusterProvisioning,
				ProviderType:          api.ClusterProviderStandalone,
				SupportedInstanceType: api.AllInstanceTypeSupport.String(),
			},
			wantErr: false,
		},
		{
			name: "should return an error if ProviderType is standalone and no ClusterDNS is set",
			input: `
---
cluster_dns: "test"
cluster_id: "test"
cloud_provider: "aws"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
status: "ready"
provider_type: "standalone"
supported_instance_type: "developer"
`,
			output: ManualCluster{
				ClusterDNS:            "test",
				ClusterId:             "test",
				CloudProvider:         "aws",
				Region:                "east-1",
				MultiAZ:               true,
				Schedulable:           true,
				KafkaInstanceLimit:    1,
				Status:                api.ClusterReady,
				ProviderType:          api.ClusterProviderStandalone,
				SupportedInstanceType: api.DeveloperTypeSupport.String(),
			},
			wantErr: true,
		},
		{
			name: "should use the provided value if they are set",
			input: `
---
name: "test"
cluster_id: "test"
cloud_provider: "aws"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
status: "ready"
provider_type: "aws_eks"
supported_instance_type: "developer"
`,
			output: ManualCluster{
				Name:                  "test",
				ClusterId:             "test",
				CloudProvider:         "aws",
				Region:                "east-1",
				MultiAZ:               true,
				Schedulable:           true,
				KafkaInstanceLimit:    1,
				Status:                api.ClusterReady,
				ProviderType:          api.ClusterProviderAwsEKS,
				SupportedInstanceType: api.DeveloperTypeSupport.String(),
			},
			wantErr: false,
		},
		{
			name: "should return error because invalid status value",
			input: `
---
name: "test"
cluster_id: "test"
cloud_provider: "aws"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
status: "not_valid"
provider_type: "aws_eks"
`,
			output:  ManualCluster{},
			wantErr: true,
		},
		{
			name: "should return error because invalid provider_type value",
			input: `
---
name: "test"
cluster_id: "test"
cloud_provider: "aws"
region: "east-1"
multi_az: true
schedulable: true
kafka_instance_limit: 1
status: "ready"
provider_type: "invalid"
`,
			output:  ManualCluster{},
			wantErr: true,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			var v ManualCluster
			err := yaml.Unmarshal([]byte(tt.input), &v)
			if err != nil && !tt.wantErr {
				t.Errorf("unexpected error %v", err)
			}
			g.Expect(v).To(gomega.Equal(tt.output))
		})
	}
}

func TestFindClusterNameByClusterId(t *testing.T) {
	type fields struct {
		ClusterList ClusterList
	}
	type args struct {
		clusterId string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "return empty when clusterId does not exist",
			fields: fields{
				ClusterList{
					ManualCluster{
						ClusterId: "12",
						Name:      "sturgis",
					},
				},
			},
			args: args{
				clusterId: "123",
			},
			want: "",
		},
		{
			name: "return correct cluster name when clusterId does exist",
			fields: fields{
				ClusterList{
					ManualCluster{
						ClusterId: "123",
						Name:      "sturgis",
					},
				},
			},
			args: args{
				clusterId: "123",
			},
			want: "sturgis",
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			clusterConfig := NewClusterConfig(tt.fields.ClusterList)
			dataplaneClusterConfig := &DataplaneClusterConfig{
				ClusterConfig: clusterConfig,
			}
			g.Expect(dataplaneClusterConfig.FindClusterNameByClusterId(tt.args.clusterId)).To(gomega.Equal(tt.want))
		})
	}
}

func TestValidateClusterIsInKubeContext(t *testing.T) {
	type args struct {
		rawKubeconfig clientcmdapi.Config
		cluster       ManualCluster
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "return error if cluster is not in kubeconfig context ",
			args: args{
				rawKubeconfig: clientcmdapi.Config{
					Contexts: map[string]*clientcmdapi.Context{},
				},
				cluster: ManualCluster{
					Name:      "glen",
					ClusterId: "123",
				},
			},
			wantErr: true,
		},
		{
			name: "return nil if cluster is in kubeconfig context ",
			args: args{
				rawKubeconfig: clientcmdapi.Config{
					Contexts: map[string]*clientcmdapi.Context{
						"glen": {},
					},
				},
				cluster: ManualCluster{
					Name:      "glen",
					ClusterId: "123",
				},
			},
			wantErr: false,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			err := validateClusterIsInKubeconfigContext(tt.args.rawKubeconfig, tt.args.cluster)
			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
		})
	}
}

func Test_GetClusterSupportedInstanceType(t *testing.T) {
	type fields struct {
		ClusterList ClusterList
	}
	type args struct {
		clusterId string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		want         bool
		wantInstType string
	}{
		{
			name: "should return standard instance type and true",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: true, SupportedInstanceType: types.STANDARD.String()},
				},
			},
			args: args{
				clusterId: "test01",
			},
			want:         true,
			wantInstType: types.STANDARD.String(),
		},
		{
			name: "should return empty string for instance type and true",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: true},
				},
			},
			args: args{
				clusterId: "test01",
			},
			want:         true,
			wantInstType: "",
		},
		{
			name: "should return empty string for instance type and false for not found cluster",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: true},
				},
			},
			args: args{
				clusterId: "test02",
			},
			want:         false,
			wantInstType: "",
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := NewClusterConfig(tt.fields.ClusterList)
			instType, found := conf.GetClusterSupportedInstanceType(tt.args.clusterId)
			g.Expect(instType).To(gomega.Equal(tt.wantInstType))
			g.Expect(found).To(gomega.Equal(tt.want))
		})
	}
}

func Test_GetManualClusters(t *testing.T) {
	type fields struct {
		ClusterList ClusterList
	}
	type args struct {
		clusterId string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []ManualCluster
	}{
		{
			name: "should return manual clusters slice from config",
			fields: fields{
				ClusterList: ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: true},
				},
			},
			args: args{
				clusterId: "test01",
			},
			want: ClusterList{
				ManualCluster{ClusterId: "test01", Schedulable: true},
			},
		},
	}

	for _, testcase := range tests {
		tt := testcase

		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := NewClusterConfig(tt.fields.ClusterList)
			g.Expect(conf.GetManualClusters()).To(gomega.Equal(tt.want))
		})
	}
}

func Test_IsReadyDataPlaneClustersReconcileEnabled(t *testing.T) {
	tests := []struct {
		name                                  string
		enableReadyDataPlaneClustersReconcile bool
		want                                  bool
	}{
		{
			name:                                  "should return true if EnableReadyDataPlaneClustersReconcile is set to true",
			enableReadyDataPlaneClustersReconcile: true,
			want:                                  true,
		},
		{
			name:                                  "should return false if EnableReadyDataPlaneClustersReconcile is set to false",
			enableReadyDataPlaneClustersReconcile: false,
			want:                                  false,
		},
	}

	for _, testcase := range tests {
		tt := testcase

		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			conf := NewDataplaneClusterConfig()
			conf.EnableReadyDataPlaneClustersReconcile = tt.enableReadyDataPlaneClustersReconcile
			g.Expect(conf.IsReadyDataPlaneClustersReconcileEnabled()).To(gomega.Equal(tt.want))
		})
	}
}

func Test_DataPlaneClusterConfig_DefaultComputeMachinesConfig(t *testing.T) {
	t.Parallel()
	type fields struct {
		config *DataplaneClusterConfig
	}
	type args struct {
		id cloudproviders.CloudProviderID
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ComputeMachinesConfig
		wantErr bool
	}{
		{
			name: "should return the provider's compute machine configuration when the provided cloud provider is defined as part of the cloud providers compute machine configuration",
			fields: fields{
				&DataplaneClusterConfig{
					DynamicScalingConfig: DynamicScalingConfig{
						ComputeMachinePerCloudProvider: map[cloudproviders.CloudProviderID]ComputeMachinesConfig{
							cloudproviders.AWS: {
								ClusterWideWorkload: &ComputeMachineConfig{
									ComputeMachineType: defaultAWSComputeMachineType,
								},
							},
						},
					},
				},
			},
			args: args{
				id: cloudproviders.AWS,
			},
			want: ComputeMachinesConfig{
				ClusterWideWorkload: &ComputeMachineConfig{
					ComputeMachineType: defaultAWSComputeMachineType,
				},
			},
			wantErr: false,
		},
		{
			name: "should return an error when the provided cloud provider is not defined as part of the cloud providers compute machine configuration",
			fields: fields{
				&DataplaneClusterConfig{
					DynamicScalingConfig: DynamicScalingConfig{
						ComputeMachinePerCloudProvider: map[cloudproviders.CloudProviderID]ComputeMachinesConfig{
							cloudproviders.GCP: {
								ClusterWideWorkload: &ComputeMachineConfig{
									ComputeMachineType: defaultGCPComputeMachineType,
								},
							},
						},
					},
				},
			},
			args: args{
				id: cloudproviders.AWS,
			},
			want:    ComputeMachinesConfig{},
			wantErr: true,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			t.Parallel()
			res, err := tt.fields.config.DefaultComputeMachinesConfig(tt.args.id)
			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
			g.Expect(res).To(gomega.Equal(tt.want))
		})
	}

}

func Test_ReadFiles(t *testing.T) {
	type fields struct {
		config *DataplaneClusterConfig
	}

	tests := []struct {
		name     string
		fields   fields
		modifyFn func(config *DataplaneClusterConfig)
		wantErr  bool
	}{
		{
			name: "should not return an error when specified ImagePullDockerConfig file does not exist",
			fields: fields{
				config: NewDataplaneClusterConfig(),
			},
			modifyFn: func(config *DataplaneClusterConfig) {
				config.ImagePullDockerConfigFile = "invalid"
				config.ImagePullDockerConfigContent = ""
			},
			wantErr: false,
		},
		{
			name: "should return an error with invalid DataPlaneClusterConfigFile",
			fields: fields{
				config: NewDataplaneClusterConfig(),
			},
			modifyFn: func(config *DataplaneClusterConfig) {
				config.DataPlaneClusterConfigFile = "invalid"
			},
			wantErr: true,
		},
		{
			name: "should not return an error and ignore non standalone provider type cluster",
			fields: fields{
				config: NewDataplaneClusterConfig(),
			},
			modifyFn: func(config *DataplaneClusterConfig) {
				config.ClusterConfig.clusterList = ClusterList{
					ManualCluster{ClusterId: "test01", Schedulable: true, ProviderType: "unknown"},
				}
			},
			wantErr: false,
		},
		{
			name: "should return an error with invalid ReadOnlyUserListFile",
			fields: fields{
				config: NewDataplaneClusterConfig(),
			},
			modifyFn: func(config *DataplaneClusterConfig) {
				config.ReadOnlyUserListFile = "invalid"
			},
			wantErr: true,
		},
		{
			name: "should return an error with invalid KafkaSREUsersFile",
			fields: fields{
				config: NewDataplaneClusterConfig(),
			},
			modifyFn: func(config *DataplaneClusterConfig) {
				config.KafkaSREUsersFile = "invalid"
			},
			wantErr: true,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			config := tt.fields.config
			if tt.modifyFn != nil {
				tt.modifyFn(config)
			}
			err := config.ReadFiles()
			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
		})
	}
}

func Test_readKubeconfig(t *testing.T) {
	type fields struct {
		config *DataplaneClusterConfig
	}

	tests := []struct {
		name     string
		fields   fields
		modifyFn func(config *DataplaneClusterConfig)
		wantErr  bool
	}{
		{
			name: "should return an error with invalid Kubeconfig",
			fields: fields{
				config: NewDataplaneClusterConfig(),
			},
			modifyFn: func(config *DataplaneClusterConfig) {
				config.Kubeconfig = "invalid"
			},
			wantErr: true,
		},
	}

	for _, testcase := range tests {
		tt := testcase
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			config := tt.fields.config
			if tt.modifyFn != nil {
				tt.modifyFn(config)
			}
			err := config.readKubeconfig()
			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
		})
	}
}
