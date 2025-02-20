package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestClusterMachineConfigRefs(t *testing.T) {
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.Kube119,
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3,
				Endpoint: &v1alpha1.Endpoint{
					Host: "test-ip",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 3,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test-1",
					},
				},
				{
					Count: 3,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test-2",
					},
				},
				{
					Count: 5,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test", // This tests duplicates
					},
				},
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test-etcd",
				},
			},
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: "eksa-unit-test",
			},
		},
	}

	want := []v1alpha1.Ref{
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test",
		},
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test-1",
		},
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test-2",
		},
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test-etcd",
		},
	}

	got := cluster.MachineConfigRefs()

	if !v1alpha1.RefSliceEqual(got, want) {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

func TestClusterIsSelfManaged(t *testing.T) {
	testCases := []struct {
		testName string
		cluster  *v1alpha1.Cluster
		want     bool
	}{
		{
			testName: "empty name",
			cluster:  &v1alpha1.Cluster{},
			want:     true,
		},
		{
			testName: "name same as self",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-1",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: "cluster-1",
					},
				},
			},
			want: true,
		},
		{
			testName: "name different tha self",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-2",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: "cluster-1",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster.IsSelfManaged()).To(Equal(tt.want))
		})
	}
}

func TestClusterSetManagedBy(t *testing.T) {
	c := &v1alpha1.Cluster{}
	managementClusterName := "managament-cluster"
	c.SetManagedBy(managementClusterName)

	g := NewWithT(t)
	g.Expect(c.IsSelfManaged()).To(BeFalse())
	g.Expect(c.ManagedBy()).To(Equal(managementClusterName))
}

func TestClusterSetSelfManaged(t *testing.T) {
	c := &v1alpha1.Cluster{}
	c.SetSelfManaged()

	g := NewWithT(t)
	g.Expect(c.IsSelfManaged()).To(BeTrue())
}

func TestClusterManagementClusterEqual(t *testing.T) {
	testCases := []struct {
		testName                                 string
		cluster1SelfManaged, cluster2SelfManaged bool
		want                                     bool
	}{
		{
			testName:            "both self managed",
			cluster1SelfManaged: true,
			cluster2SelfManaged: true,
			want:                true,
		},
		{
			testName:            "both managed",
			cluster1SelfManaged: false,
			cluster2SelfManaged: false,
			want:                true,
		},
		{
			testName:            "one managed, one self managed",
			cluster1SelfManaged: false,
			cluster2SelfManaged: true,
			want:                false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{}
			setSelfManaged(cluster1, tt.cluster1SelfManaged)
			cluster2 := &v1alpha1.Cluster{}
			setSelfManaged(cluster2, tt.cluster2SelfManaged)

			g := NewWithT(t)
			g.Expect(cluster1.ManagementClusterEqual(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualKubernetesVersion(t *testing.T) {
	testCases := []struct {
		testName                         string
		cluster1Version, cluster2Version v1alpha1.KubernetesVersion
		want                             bool
	}{
		{
			testName:        "both empty",
			cluster1Version: "",
			cluster2Version: "",
			want:            true,
		},
		{
			testName:        "one empty, one exists",
			cluster1Version: "",
			cluster2Version: v1alpha1.Kube118,
			want:            false,
		},
		{
			testName:        "both exists, diff",
			cluster1Version: v1alpha1.Kube118,
			cluster2Version: v1alpha1.Kube119,
			want:            false,
		},
		{
			testName:        "both exists, same",
			cluster1Version: v1alpha1.Kube118,
			cluster2Version: v1alpha1.Kube118,
			want:            true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: tt.cluster1Version,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: tt.cluster2Version,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualWorkerNodeGroupConfigurations(t *testing.T) {
	testCases := []struct {
		testName                   string
		cluster1Wngs, cluster2Wngs []v1alpha1.WorkerNodeGroupConfiguration
		want                       bool
	}{
		{
			testName:     "both empty",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{},
			want:         true,
		},
		{
			testName: "one empty, one exists",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 1,
				},
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 1,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k",
						Name: "n",
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 1,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k",
						Name: "n",
					},
				},
			},
			want: true,
		},
		{
			testName: "both exist, order diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 1,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k1",
						Name: "n1",
					},
				},
				{
					Count: 2,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k2",
						Name: "n2",
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 2,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k2",
						Name: "n2",
					},
				},
				{
					Count: 1,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k1",
						Name: "n1",
					},
				},
			},
			want: true,
		},
		{
			testName: "both exist, count diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 1,
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: 2,
				},
			},
			want: false,
		},
		{
			testName: "both exist, ref diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k1",
						Name: "n1",
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k2",
						Name: "n2",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					WorkerNodeGroupConfigurations: tt.cluster1Wngs,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					WorkerNodeGroupConfigurations: tt.cluster2Wngs,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualDatacenterRef(t *testing.T) {
	testCases := []struct {
		testName                                     string
		cluster1DatacenterRef, cluster2DatacenterRef v1alpha1.Ref
		want                                         bool
	}{
		{
			testName: "both empty",
			want:     true,
		},
		{
			testName: "one empty, one exists",
			cluster1DatacenterRef: v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			want: false,
		},
		{
			testName: "both exist, diff",
			cluster1DatacenterRef: v1alpha1.Ref{
				Kind: "k1",
				Name: "n1",
			},
			cluster2DatacenterRef: v1alpha1.Ref{
				Kind: "k2",
				Name: "n2",
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1DatacenterRef: v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			cluster2DatacenterRef: v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					DatacenterRef: tt.cluster1DatacenterRef,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					DatacenterRef: tt.cluster2DatacenterRef,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualIdentityProviderRefs(t *testing.T) {
	testCases := []struct {
		testName                 string
		cluster1Ipr, cluster2Ipr []v1alpha1.Ref
		want                     bool
	}{
		{
			testName:    "both empty",
			cluster1Ipr: []v1alpha1.Ref{},
			cluster2Ipr: []v1alpha1.Ref{},
			want:        true,
		},
		{
			testName: "one empty, one exists",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k",
					Name: "n",
				},
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Ipr: []v1alpha1.Ref{
				{
					Kind: "k",
					Name: "n",
				},
			},
			want: true,
		},
		{
			testName: "both exist, order diff",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k1",
					Name: "n1",
				},
				{
					Kind: "k2",
					Name: "n2",
				},
			},
			cluster2Ipr: []v1alpha1.Ref{
				{
					Kind: "k2",
					Name: "n2",
				},
				{
					Kind: "k1",
					Name: "n1",
				},
			},
			want: true,
		},
		{
			testName: "both exist, count diff",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k1",
					Name: "n1",
				},
			},
			cluster2Ipr: []v1alpha1.Ref{
				{
					Kind: "k2",
					Name: "n2",
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					IdentityProviderRefs: tt.cluster1Ipr,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					IdentityProviderRefs: tt.cluster2Ipr,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualGitOpsRef(t *testing.T) {
	testCases := []struct {
		testName                             string
		cluster1GitOpsRef, cluster2GitOpsRef *v1alpha1.Ref
		want                                 bool
	}{
		{
			testName:          "both nil",
			cluster1GitOpsRef: nil,
			cluster2GitOpsRef: nil,
			want:              true,
		},
		{
			testName: "one nil, one exists",
			cluster1GitOpsRef: &v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			cluster2GitOpsRef: nil,
			want:              false,
		},
		{
			testName: "both exist, diff",
			cluster1GitOpsRef: &v1alpha1.Ref{
				Kind: "k1",
				Name: "n1",
			},
			cluster2GitOpsRef: &v1alpha1.Ref{
				Kind: "k2",
				Name: "n2",
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1GitOpsRef: &v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			cluster2GitOpsRef: &v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					GitOpsRef: tt.cluster1GitOpsRef,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					GitOpsRef: tt.cluster2GitOpsRef,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualClusterNetwork(t *testing.T) {
	testCases := []struct {
		testName                                       string
		cluster1ClusterNetwork, cluster2ClusterNetwork v1alpha1.ClusterNetwork
		want                                           bool
	}{
		{
			testName:               "both nil",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{},
			want:                   true,
		},
		{
			testName: "one empty, one exists",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, diff",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/6",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ClusterNetwork: tt.cluster1ClusterNetwork,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ClusterNetwork: tt.cluster2ClusterNetwork,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualExternalEtcdConfiguration(t *testing.T) {
	testCases := []struct {
		testName                   string
		cluster1Etcd, cluster2Etcd *v1alpha1.ExternalEtcdConfiguration
		want                       bool
	}{
		{
			testName:     "both nil",
			cluster1Etcd: nil,
			cluster2Etcd: nil,
			want:         true,
		},
		{
			testName: "one nil, one exists",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Etcd: nil,
			want:         false,
		},
		{
			testName: "both exist, same",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			want: true,
		},
		{
			testName: "both exist, count diff",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 2,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			want: false,
		},
		{
			testName: "both exist, ref diff",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k1",
					Name: "n1",
				},
			},
			cluster2Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k2",
					Name: "n2",
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ExternalEtcdConfiguration: tt.cluster1Etcd,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ExternalEtcdConfiguration: tt.cluster2Etcd,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualProxyConfiguration(t *testing.T) {
	testCases := []struct {
		testName                     string
		cluster1Proxy, cluster2Proxy *v1alpha1.ProxyConfiguration
		want                         bool
	}{
		{
			testName:      "both nil",
			cluster1Proxy: nil,
			cluster2Proxy: nil,
			want:          true,
		},
		{
			testName: "one nil, one exists",
			cluster1Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			cluster2Proxy: nil,
			want:          false,
		},
		{
			testName: "both exist, same",
			cluster1Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			cluster2Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			want: true,
		},
		{
			testName: "both exist, diff",
			cluster1Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			cluster2Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.5",
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProxyConfiguration: tt.cluster1Proxy,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProxyConfiguration: tt.cluster2Proxy,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualRegistryMirrorConfiguration(t *testing.T) {
	testCases := []struct {
		testName                   string
		cluster1Regi, cluster2Regi *v1alpha1.RegistryMirrorConfiguration
		want                       bool
	}{
		{
			testName:     "both nil",
			cluster1Regi: nil,
			cluster2Regi: nil,
			want:         true,
		},
		{
			testName: "one nil, one exists",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: nil,
			want:         false,
		},
		{
			testName: "both exist, same",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			want: true,
		},
		{
			testName: "both exist, diff",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.5",
				CACertContent: "ca",
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: tt.cluster1Regi,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: tt.cluster2Regi,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualManagement(t *testing.T) {
	testCases := []struct {
		testName                               string
		cluster1Management, cluster2Management string
		want                                   bool
	}{
		{
			testName:           "both empty",
			cluster1Management: "",
			cluster2Management: "",
			want:               true,
		},
		{
			testName:           "one empty, one equal to self",
			cluster1Management: "",
			cluster2Management: "cluster-1",
			want:               true,
		},
		{
			testName:           "both equal to self",
			cluster1Management: "cluster-1",
			cluster2Management: "cluster-1",
			want:               true,
		},
		{
			testName:           "one empty, one not equal to self",
			cluster1Management: "",
			cluster2Management: "cluster-2",
			want:               false,
		},
		{
			testName:           "one equal to self, one not equal to self",
			cluster1Management: "cluster-1",
			cluster2Management: "cluster-2",
			want:               false,
		},
		{
			testName:           "both not equal to self and different",
			cluster1Management: "cluster-2",
			cluster2Management: "cluster-3",
			want:               false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-1",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: tt.cluster1Management,
					},
				},
			}
			cluster2 := &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-1",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: tt.cluster2Management,
					},
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestControlPlaneConfigurationEqual(t *testing.T) {
	testCases := []struct {
		testName                           string
		cluster1CPConfig, cluster2CPConfig *v1alpha1.ControlPlaneConfiguration
		want                               bool
	}{
		{
			testName: "both exist, same",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			want: true,
		},
		{
			testName: "one nil, one exists",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2CPConfig: nil,
			want:             false,
		},
		{
			testName: "one nil, one exists",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2CPConfig: nil,
			want:             false,
		},
		{
			testName: "count exists, diff",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 2,
			},
			want: false,
		},
		{
			testName: "one count empty",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{},
			want:             false,
		},
		{
			testName: "endpoint diff",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.5",
				},
			},
			want: false,
		},
		{
			testName: "one endpoint empty",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Endpoint: nil,
			},
			want: false,
		},
		{
			testName: "both endpoints empty",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "",
				},
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{},
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster1CPConfig.Equal(tt.cluster2CPConfig)).To(Equal(tt.want))
		})
	}
}

func TestRegistryMirrorConfigurationEqual(t *testing.T) {
	testCases := []struct {
		testName                   string
		cluster1Regi, cluster2Regi *v1alpha1.RegistryMirrorConfiguration
		want                       bool
	}{
		{
			testName:     "both nil",
			cluster1Regi: nil,
			cluster2Regi: nil,
			want:         true,
		},
		{
			testName: "one nil, one exists",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: nil,
			want:         false,
		},
		{
			testName: "both exist, same",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			want: true,
		},
		{
			testName: "both exist, endpoint diff",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint: "1.2.3.5",
			},
			want: false,
		},
		{
			testName: "both exist, ca diff",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				CACertContent: "ca1",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				CACertContent: "ca2",
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster1Regi.Equal(tt.cluster2Regi)).To(Equal(tt.want))
		})
	}
}

func setSelfManaged(c *v1alpha1.Cluster, s bool) {
	if s {
		c.SetSelfManaged()
	} else {
		c.SetManagedBy("management-cluster")
	}
}
