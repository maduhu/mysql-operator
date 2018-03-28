package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// MysqlClusterKind is the kind of the crd that will be created.
	MysqlClusterKind = "MysqlCluster"
	// MysqlClusterPlural is the plural for mysqlcluster
	MysqlClusterPlural = "mysqlclusters"
	groupName          = "mysql.presslabs.net"

	MysqlBackupKind   = "MysqlBackup"
	MysqlBackupPlural = "mysqlbackups"
)

var (
	// SchemeBuilder the scheme builder
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme function
	AddToScheme = SchemeBuilder.AddToScheme
	// SchemeGroupVersion ..
	SchemeGroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha1"}
	// MysqlClusterCRDName the crd name
	MysqlClusterCRDName = MysqlClusterPlural + "." + groupName

	// MysqlBackupCRDName the crd name of backup resource
	MysqlBackupCRDName = MysqlBackupPlural + "." + groupName
)

// Resource gets an MysqlCluster GroupResource for a specified resource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion,
		&MysqlCluster{},
		&MysqlClusterList{},
		&MysqlBackup{},
		&MysqlBackupList{},
	)
	metav1.AddToGroupVersion(s, SchemeGroupVersion)
	return nil
}