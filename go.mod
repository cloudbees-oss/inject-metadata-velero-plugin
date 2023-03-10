module github.com/cloudbees/inject-metadata-velero-plugin

go 1.14

require (
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/vmware-tanzu/velero v1.10.2
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
)

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
