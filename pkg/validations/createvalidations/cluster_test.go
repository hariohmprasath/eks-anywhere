package createvalidations_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
)

const testclustername string = "testcluster"

func TestValidateClusterPresent(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "SuccessNoClusters",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_cluster_response.json",
		},
		{
			name:               "FailureClusterNameExists",
			wantErr:            errors.New("cluster name testcluster already exists"),
			getClusterResponse: "testdata/cluster_name_exists.json",
		},
		{
			name:               "SuccessClusterNotInList",
			wantErr:            nil,
			getClusterResponse: "testdata/name_not_in_list.json",
		},
	}

	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(ctx, []string{"get", capiClustersResourceType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}).Return(*bytes.NewBufferString(fileContent), nil)
			err := createvalidations.ValidateClusterNameIsUnique(ctx, k, cluster, testclustername)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateManagementClusterCRDs(t *testing.T) {
	tests := []struct {
		name                      string
		wantErr                   bool
		errGetClusterCRD          error
		errGetClusterCRDCount     int
		errGetEKSAClusterCRD      error
		errGetEKSAClusterCRDCount int
	}{
		{
			name:                      "Success",
			wantErr:                   false,
			errGetClusterCRD:          nil,
			errGetClusterCRDCount:     1,
			errGetEKSAClusterCRD:      nil,
			errGetEKSAClusterCRDCount: 1,
		},
		{
			name:                      "FailureClusterCRDDoesNotExist",
			wantErr:                   true,
			errGetClusterCRD:          errors.New("cluster CRD does not exist"),
			errGetClusterCRDCount:     1,
			errGetEKSAClusterCRD:      nil,
			errGetEKSAClusterCRDCount: 0,
		},
		{
			name:                      "FailureEKSAClusterCRDDoesNotExist",
			wantErr:                   true,
			errGetClusterCRD:          nil,
			errGetClusterCRDCount:     1,
			errGetEKSAClusterCRD:      errors.New("eksa cluster CRDS do not exist"),
			errGetEKSAClusterCRDCount: 1,
		},
	}

	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			e.EXPECT().Execute(ctx, []string{"get", "crd", capiClustersResourceType, "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, tc.errGetClusterCRD).Times(tc.errGetClusterCRDCount)
			e.EXPECT().Execute(ctx, []string{"get", "crd", eksaClusterResourceType, "--kubeconfig", cluster.KubeconfigFile}).Return(bytes.Buffer{}, tc.errGetEKSAClusterCRD).Times(tc.errGetEKSAClusterCRDCount)
			err := createvalidations.ValidateManagementCluster(ctx, k, cluster)
			if tc.wantErr {
				assert.Error(tt, err, "expected ValidateManagementCluster to return an error", "test", tc.name)
			} else {
				assert.NoError(tt, err, "expected ValidateManagementCluster not to return an error", "test", tc.name)
			}
		})
	}
}

var (
	capiClustersResourceType = fmt.Sprintf("clusters.%s", v1alpha3.GroupVersion.Group)
	eksaClusterResourceType  = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
)
