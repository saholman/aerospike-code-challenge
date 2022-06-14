# Aerospike Code Challenge

## Exercise Description
Create a program in golang that interacts with a k8s cluster using the client-go library (GitHub - kubernetes/client-go: Go client for Kubernetes. ).  The program should perform the following:

* connect to the k8s cluster
* print out the namespaces on the cluster
* create a new namespace
* create a pod in that namespace that runs a simple hello-world container
* print out pod names and the namespace they are in for any pods that have a label of ‘k8s-app=kube-dns’ or a similar label is ok as well
* delete the hello-world pod created from above
* extra credit - show how an client-go informer works

The example should be loaded into a github repo of the candidate’s choice to assist in reviewing of the code.  The candidate should be able to describe the following:
* how they set up their k8s dev host
* what tools they used to code up the example
* how their code is structured and what it does including how they used features of the client-go library

