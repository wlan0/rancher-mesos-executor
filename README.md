# Rancher Mesos Executor

This repository contains the executor of the apache mesos framework for integrating rancher with mesos.

The executor will spin up a new VM on the Mesos-slave using libvirt, and install docker on this new VM, followed by registering it with rancher. The slave to run on is chosen by mesos. Rancher provides it private networking, load balancing and service discovery to the newly created VMs.

# Contact
For bugs, questions, comments, corrections, suggestions, etc., open an issue in
 [rancher/rancher](//github.com/rancher/rancher/issues) with a title starting with `[rancher-mesos-executor] `.

 Or just [click here](//github.com/rancher/rancher/issues/new?title=%5Brancher-mesos-executor%5D%20) to create a new issue.

# License
Copyright (c) 2014-2015 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
