import random
import struct
import sys
sys.path.append('/home/ubuntu/FaaSGraph/test')
import config_manager
import time
import requests
import json

def int2byte(a: int) -> bytes:
    return struct.pack('i', a)

def byte2int(byte: bytes) -> int:
    return struct.unpack('i', byte)[0]

def byte2float(byte: bytes) -> float:
    return struct.unpack('f', byte)[0]

# def build():
#     global startNodeSet
#     with open(data_path + '/start_nodes.bin', 'rb') as f:
#         for _ in range(partitions):
#             startNodeSet.append(byte2int(f.read(4)))
#         startNodeSet.append(numV)

def run(graph, url, app, function, source=0):
    request_id = str(random.randint(1, 10000000))
    data = {'graph': graph, 'request_id': request_id, 'app': app, 'function': function, 'in_edge': False, 'out_edge': True, 'param': {'source': source}}
    result = requests.post(url, json=data)
    return json.loads(result.content)

# def gather():
#     results = {}
#     for i in range(len(startNodeSet) - 1):
#         startNode = startNodeSet[i]
#         size = startNodeSet[i+1] - startNodeSet[i]
#         with open('{}/{}/vertex_val.bin'.format(config_manager.config['DATA_PATH'], startNode), 'rb') as f:
#             for v in range(startNode, startNode + size):
#                 results[v] = byte2float(f.read(4))
#     with open('serverless_bfs_result.txt', mode='w') as f:
#         results = sorted(results.items(), key=lambda kv: kv[0])
#         for node, rank in results:
#             f.write('{}: {}\n'.format(node, rank))

if __name__ == '__main__':
    graph, source = sys.argv[1], 0
    if len(sys.argv) == 3:
        source = int(sys.argv[2])
    partitions = config_manager.config['GRAPH'][graph]['PARTITION']
    numV = config_manager.config['GRAPH'][graph]['TOTAL_NODE']
    data_path = config_manager.config['GRAPH'][graph]['DATA_PATH']
    startNodeSet = []
    # build()
    request_id = str(random.randint(1, 100000000))
    start = time.time()
    result = run(graph, 'http://localhost:20001/run_graph', 'graph', 'bfs', source)
    end = time.time()
    print(result)
    print('startup: {}'.format(result['startup']))
    print('io: {}'.format(result['io']))
    print('preprocess: {}'.format(result['preprocess']))
    print('query: {}'.format(result['query']))
    print('store: {}'.format(result['store']))
    print('comm: {}'.format(result['comm']))
    print('latency: {}'.format(end - start))
    # gather()
