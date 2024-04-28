# K8s Operator

Operator is a type of controller that implements a specific operational logic to manage a group of k8s resources. Operator makes use of the control loop to manage the resources’ state. The control loop logic brings the state of the target resource to from the current state to the desired state

# Features
#### Streamlined Partition Deletion:
Efficiently manage the deletion of partitions within DynamoDB tables using Kubernetes Custom Resources, enabling seamless declarative configuration.

#### Kubebuilder Integration:
Leverage the Kubebuilder framework to ensure adherence to industry best practices, facilitating the development of scalable and robust Kubernetes operators.

#### Self-Healing Mechanisms:
Operators can monitor the state of applications and automatically perform actions to restore them to the desired state when issues arise, thus enhancing the system's self-healing capabilities.

## Custom resource

CRD — Custom Resource Definition — defines a Custom Resource which is not available in the default Kubernetes implementation.

Example Custom resource to be used for the delete partition (DP) job:

```yaml
apiVersion: dynamoctl.dp.operators.sanjivmadhavan.io/v1alpha1
kind: DdbctlDpJob
metadata:
  name: ddbctldpjob-sample
spec:
  ddbCtlDpJob:
    awsRegion: us-east-1
    endpointURL: http://aws-dynamhttp://dynamodb.local:8000
    partitionValue: partition-key-value
    tableName: my-dynamodb-table
```

The dynamoctl-dp-operator extends the operator pattern of k8s using Custom resources to specify the intent to delete a partition from a dynamoDB table. The job is based ioff on helm charts. We use kubebuilder to scaffold the project

## Kubebuilder
Kubebuilder is a framework for building Kubernetes APIs / Operators, which helps to generate a set of boiler plate codes for the Controller, and related CRDs.

#### Command references
To scaffold the project, use:

```bash kubebuilder init --domain dynamoctl.dp.operators.<name>.io --repo <your-github-repo>```

