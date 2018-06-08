.. _kops_guide:

**********************************
Kubernetes Kops Installation Guide
**********************************

Step-by-step guide to create Kubernetes cluster on AWS using kops and Cilium as CNI.

Prerequisites
=============

* aws cli
* kubectl
* aws account with permissions AmazonEC2FullAccess, AmazonRoute53FullAccess, AmazonS3FullAccess, IAMFullAccess, AmazonVPCFullAccess 

Installing KOPS
===============

 In order to control the kubernetes clusters we need to install kubectl cli. Installation details can be found by following this link. Below instructions for KOPS are for linux, please use instructions relevant for your setup.
Download the latest kops binary

.. code:: bash

        curl -LO https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
        chmod +x kops-linux-amd64
        sudo mv kops-linux-amd64 /usr/local/bin/kops

        # If you are on MacOS, you can use:
        brew update && brew install kops



Setting up KOPS
===============

Assuming you have all the pre-requisites, run the following commands to create the ``kops`` user and group

.. code:: bash

        # Create IAM group named kops and grant access
        aws iam create-group --group-name kops
        aws iam attach-group-policy --policy-arn arn:aws:iam::aws:policy/AmazonEC2FullAccess --group-name kops
        aws iam attach-group-policy --policy-arn arn:aws:iam::aws:policy/AmazonRoute53FullAccess --group-name kops
        aws iam attach-group-policy --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess --group-name kops
        aws iam attach-group-policy --policy-arn arn:aws:iam::aws:policy/IAMFullAccess --group-name kops
        aws iam attach-group-policy --policy-arn arn:aws:iam::aws:policy/AmazonVPCFullAccess --group-name kops
        aws iam create-user --user-name kops
        aws iam add-user-to-group --user-name kops --group-name kops
        aws iam create-access-key --user-name kops


kops requires a creation of a dedicated s3 bucket in order to store the state and representation of the cluster. You will need to change the bucket name and provide your unique bucket name (for example a reverse of FQDN added with short description of the cluster). Also make sure to use the region where you will be deploying the cluster. 

.. code:: bash

        aws s3api create-bucket --bucket prefix-example-com-state-store --region us-east-1
        export KOPS_STATE_STORE=s3://prefix-example-com-state-store

Above steps are sufficient for getting a working cluster installed. Please consult [kops aws documentation](https://github.com/kubernetes/kops/blob/master/docs/aws.md) for more detailed setup

Cilium Prerequisites
====================

* Cilium requires minimum kernel version to be 4.8
* Cilium requires minimum etcd version as 3.1  
* Kubernetes with CRD validation (Recommended) (more on this after initial dry run setup of the cluster)

In this guide we will use the etcd version as 3.1.11 and the latest CoreOS stable image  which satisfies the minimum kernel version requirement of cilium. To get the latest CoreOS ami image, you can change the region value to your choice in the command below.

.. code:: bash
        
        aws ec2 describe-images --region=us-east-1 --owner=595879546273 --filters "Name=virtualization-type,Values=hvm" "Name=name,Values=CoreOS-stable*" --query 'sort_by(Images,&CreationDate)[-1].{id:ImageLocation}'

.. code:: json

        {
                "id": "595879546273/CoreOS-stable-1745.3.1-hvm"
        }


Creating the Cluster
====================

* Note that you will need to specify the ``--master-zones`` and ``--zones`` for creating the master and worker nodes. The number of master zones should be odd (1, 3, ...) for the HA. For simplicity, you can just use 1 region.
* the cluster ``NAME`` variable should end in ``k8s.local`` to use the gossip protocol. 

.. code:: bash

        export NAME=cilium.k8s.local
        export KOPS_FEATURE_FLAGS=SpecOverrideFlag 
        kops create cluster  --state=${KOPS_STATE_STORE}  --node-count 3 --node-size t2.medium --master-size t2.medium --topology private --master-zones eu-west-1a,eu-west-1b,eu-west-1c --zones eu-west-1a,eu-west-1b,eu-west-1c --image 595879546273/CoreOS-stable-1745.3.1-hvm --networking cilium --override "cluster.spec.etcdClusters[*].version=3.1.11" --kubernetes-version 1.10.3  --cloud-labels "Team=Dev,Owner=Admin"  ${NAME}


You may be prompted to create a ssh public-private key pair.

.. code:: bash

        ssh-keygen


(Please see appendix for details on the flags used to create the cluster.)

Kubernetes with CRD validation 
==============================

In order to enable the flag ``--feature-gates=CustomResourceValidation=true``, edit the cluster yaml

.. code:: bash
        
        kops edit cluster --name= ${NAME}

Append the below yaml snippet to the end

.. code:: YAML

  kubeAPIServer:
    featureGates:
      CustomResourceValidation: "true"


After successful editing , apply changes using update cluster. 

.. code:: bash

        kops update cluster ${NAME} --yes
        kops validate cluster


Upgrading Cilium
=================

The default Cilium version deployed by kops is old. So we need to upgrade the Cilium daemonset to a newer version. The below commands illustrate the upgrade process for the Kubernetes v1.10 since that is the version we created. And we are upgrading Cilium to ``v1.0.3`` but you can replace to any stable version ``vX.Y.Z``. (Please consult [this doc](http://cilium.readthedocs.io/en/latest/install/upgrade/) for more details on Cilium upgrade.)

.. code:: bash
        
        kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/HEAD/examples/kubernetes/1.10/cilium-rbac.yaml
        kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/HEAD/examples/kubernetes/1.10/cilium-ds.yaml
        kubectl set image daemonset/cilium -n kube-system cilium-agent=docker.io/cilium/cilium:v1.0.3
        kubectl rollout status daemonset/cilium -n kube-system

Testing Cilium
==============
Follow the [Cilium getting started guide example](http://cilium.readthedocs.io/en/latest/gettingstarted/minikube/#step-2-deploy-the-demo-application) to test the cluster is setup properly and that Cilium CNI and security policies are functional.
        

Appendix: Details of kops flags used in cluster creation
========================================================

The following section explains all the flags used in create cluster command. 

* ``KOPS_FEATURE_FLAGS=SpecOverrideFlag`` : This flag is used to override the etcd version to be used from 2.X[kops default ] to 3.1.x [requirement of cilium]
* ``--state=${KOPS_STATE_STORE}`` : KOPS uses an s3 bucket to store the state of your cluster and representation of your cluster
* ``--node-count 3`` : No. of worker nodes in the kubernetes cluster.
* ``--node-size t2.medium`` : The size of the AWS EC2 instance for worker nodes
* ``--master-size t2.medium`` : The size of the AWS EC2 instance of master nodes
* ``--topology private`` : Cluster will be created with private topology, what that means is all masters/nodes will be launched in a private subnet in the VPC
* ``--master-zones eu-west-1a,eu-west-1b,eu-west-1c`` : This ensures the HA of master nodes [3] each belonging in a different Availability zones.
* ``--zones eu-west-1a,eu-west-1b,eu-west-1c`` : Where the worker nodes will be deployed
* ``--image 595879546273/CoreOS-stable-1745.3.1-hvm`` : Image name to be deployed (Cilium requires kernel version 4.8 and above so ensure to use the right OS for workers.)
* ``--networking cilium`` : Networking CNI plugin to be used - cilium 
* ``--override "cluster.spec.etcdClusters[*].version=3.1.11"`` : Overrides the etcd version to be used.
* ``--kubernetes-version 1.10.3`` : Kubernetes version that is to be installed. Please note [Kops 1.9 officially supports k8s version 1.9]
* ``--cloud-labels "Team=Dev,Owner=Admin"`` :  Labels for your cluster
* ``${NAME}`` : Name of the cluster. Make sure the name ends with k8s.local for a gossip based cluster

	
 
