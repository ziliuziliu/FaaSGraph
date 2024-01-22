# FaaSGraph

[![Static Badge](https://img.shields.io/badge/Organization_Website-EPCC-purple)](http://epcc.sjtu.edu.cn)

## Introduction

FaaSGraph is a scalable, efficient and cost-effective graph processing framework supported by serverless computing. It serves as a graph-as-a-service solution targeted at cloud providers.

FaaSGraph employs a data-centric serverless execution model, introducing a fundamental unit (named ContainerSet) for a single heavyweight computing task. Locality aware resource fusion breaks down the barriers between containers, enabling intra-host container resource sharing and shared memory communication. Meanwhile, inter-host containers communicate via the network with computation-communication interleaving and message consolidation. The bounded resource scaling policy employs a queuing method to limit container cold start rate, minimizing memory footprint while processing diverse workloads. Furthermore, we re-evaluate design choices in message passing and graph storage, minimizing auxiliary data structures to reduce memory consumption and preprocessing overhead. 

## Hardware Configuration

FaaSGraph can be deployed in both shared-memory (single-host) and distributed (multi-host) environments. All hosts run Ubuntu 20.04.

1) We recommend each node to have at least {Cores: 24, Memory: 96GB, Disk: 100GB SSD}.

2) In serverless systems, the containers (or functions) generally pull data from remote storage. To simulate remote storage access in FaaSGraph, we suggest connecting a Network Attached Storage (NAS) to each host, with the graph dataset residing in the NAS.

## Installation

In every host, clone the code in $HOME directory, enter this project and run:

```
    cd util
    ./prepare_machine.bash
```
This will install golang, docker, python package and build necessary images.

## Download Graph Dataset and Build CSR

### Download and Build CSR for amazon, livejournal and twitter

```
cd util
./prepare_dataset.bash
```

This would take roughly an hour. Please note that this script does not include friendster because it would take more than 96GB memory to build its CSR binary data. Please follow the instruction below if you want to build CSR for friendster in a larger server.

### For Specific Dataset

Download from url, then extract the raw txt: 

```
gzip -d <DATA_TXT_GZ_PATH>
```

Run a simple hashing script to ensure that the vertex IDs start from 0:

```
cd util
go run hash.go <RAW_TXT_PATH> <HASHED_PATH> <SPLIT_CHAR> <TOTAL_VERTEX>
    -- RAW_TXT_PATH: the extracted graph dataset file path.
    -- HASHED_PATH: the hashed graph dataset file path.
    -- SPLIT_CHAR: TAB => vertex IDs are separated by tab in RAW_TXT_PATH; SPACE => vertex IDs are separated by blank space in RAW_TXT_PATH. You can use "head <RAW_TXT_PATH>" to find out the split character.
    -- TOTAL_VERTEX: total number of vertex in the dataset. You can locate this in snap information page.
```

Then build the CSR/CSC binary file: 

```
cd util
go run txt2csr.go <HASHED_PATH> <CSR_PATH> <W> <TOTAL_VERTEX>
    -- HASHED_PATH: the hashed graph dataset file path.
    -- CSR_PATH: the csr file directory.
    -- W: WEIGHTED or UNWEIGHTED. When running SSSP, this should be WEIGHTED.
    -- TOTAL_VERTEX: total number of vertex in the dataset.
```

Then, build the CSR offsets:

```
cd util
go run split.go <HASHED_PATH> <CSR_PATH> <W> <TOTAL_VERTEX> <PARTITION>
    -- PARTITION: the amount of partitions. For each <PARTITION> a different set of offsets should be constructed.
```

### URLs (Big thanks to SNAP!)

1) amazon: https://snap.stanford.edu/data/com-Amazon.html

2) livejournal: https://snap.stanford.edu/data/soc-LiveJournal1.html

3) twitter: https://snap.stanford.edu/data/twitter-2010.html

4) friendster: https://snap.stanford.edu/data/com-Friendster.html


## config/config.json - Modify them in every host

Modify the following configuration before starting servers.

```
-- CONTROLLER: the address of faas_controller.
-- WORKER: all available worker addresses that can serve graph queries.
-- MAX_SLOT_PER_NODE: the amount of CPU in each host. We assume each host has the same number of CPUs.
```

Following are the graph dataset and graph processing application metadata.

```
-- GRAPH: graph dataset metadata.
  -- DATASET_NAME: name of dataset
    -- DATA_PATH: csr file directory.
    -- WEIGHTED: true or false.
    -- TOTAL_NODE: total number of vertex in the dataset.
    -- PARTITION: the intended number of partitions. Note that we assign each graph partition a distinct CPU core.
-- APP:
  -- NAME
  -- FUNCTION:
    -- NAME: name of graph processing application
    -- MAX_SLOT_PER_CONTAINER: cores per container. Default is 2.
```

Following are the advanced settings that tunes the behavior of FaaSGraph.

```
-- QUEUE_LENGTH: adjust the length of Queue in the Bounded Scaling policy
-- SHARE_CPU: true or false
-- SHARE_MEMORY: true or false
-- CURRENT_TAKEN_SLOT: simulate the environment that certain amount of cores is already taken
```

## Basic Functionality Test

In this example, we run the basic Pagerank algorithm on top of the livejournal graph dataset in single-host environment.

1) Open a terminal, run:

```
cd src/faas_controller
go build -o faas_controller ./src
    -- This will build a binary of faas_controller server.
./faas_controller
    -- This will start the server in port 20001.
```

2) Open another terminal, run:

```
cd src/local_manager
go build -o local_manager ./src
    -- This will build a binary of local_manager server.
sudo ./local_manager
    -- This will start the server in port 20000.
```

3) Open another terminal, run:

```
cd test/pr
python3 run.py livejournal-unweighted
```

The script will invoke a graph processing query. After a while, the script returns with some performance metrics. The most important ones include:

```
-- startup: container start-up time
-- io: the time to load graph dataset from storage
-- preprocess: preprocessing time to build additional data structures
-- query: the time to execute the graph processing algorithm
-- store: the time to store the results
-- latency: end-to-end graph processing latency
```

## Run General Experiment

Run: 

```
cd experiment
sudo python3 run.py
```

This will take roughly an hour. This script reproduces the performance metrics presented in Table 3 and Figure 7 of the FaaSGraph paper (latency breakdown of FaaSGraph). All cases are runnable in a single node. It initiates the faas_controller server and local_manager server, then invokes graph processing queries using a variety of datasets (amazon, livejournal, twitter) and graph applications (bfs, cc, pr, sssp). The anticipated outcome is stored in the "experiment/expected_result" directory. Please note that IO performance are influenced by the choice of storage (local file system or NAS), leading to potential latency variations.

## Distributed Deployment

Suppose you have a cluster of servers (host A, B, C, D). Modify the ```config/config.json``` profile, and:

1) Pick one host to run a faas_controller server (host A).

2) In every host, run a local_manager server (host A, B, C, D).

3) In host A, run:

```
cd test/<APP>
python3 run.py <GRAPH>
```

## Graph Processing Query Workoad

Unfortunately, we are unable to disclose additional details regarding the query workload illustrated in Figure 1 from the FaaSGraph paper due to regulatory constraints. However, we recommend a comparable query trace available at https://www.kaggle.com/datasets/chicago/chicago-taxi-trips-bq. This dataset consists of taxi trip information in Chicago and exhibits a temporal distribution throughout the day that closely resembles the patterns observed in our research. We are still actively pursuing the necessary approvals to open-source the query workload used in the paper.

## Cite

TO_BE_APPEAR