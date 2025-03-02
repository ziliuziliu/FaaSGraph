import gevent
from gevent import monkey
monkey.patch_all()

import pandas
import subprocess
import time
from typing import List
import os
import numpy
import json
import re

FAASGRAPH_CONTROLLER_PATH = '/home/ubuntu/FaaSGraph/src/faas_controller/faas_controller'
LOCAL_MANAGER_PATH = '/home/ubuntu/FaaSGraph/src/local_manager/local_manager'
APPLICAITON_PATH = '/home/ubuntu/FaaSGraph/test/{}/run.py'

mem_end_flag = False
basic_mem, max_mem = 0, 0

def get_mem(type, remote_node = None):
    if type == 'local':
        ret = subprocess.run('free | grep Mem', shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        result_str = ret.stdout.decode('utf-8', 'ignore').splitlines()
        return float(result_str[0].split()[2]) / 1024 / 1024
    else:
        return 0

def start_mem_monitor(remote_nodes):
    global basic_mem, max_mem
    basic_mem += get_mem('local')
    for remote_node in remote_nodes:
        basic_mem += get_mem('remote', remote_node)
    max_mem = basic_mem
    while True:
        if mem_end_flag == True:
            break
        mem = get_mem('local')
        for remote_node in remote_nodes:
            mem += get_mem('remote', remote_node)
        if mem > max_mem:
            max_mem = mem
        gevent.sleep(0.1)

def run_faasgraph(dataset, app):
    global mem_end_flag, basic_mem, max_mem
    mem_end_flag, basic_mem, max_mem = False, 0, 0
    mem_job = gevent.spawn(start_mem_monitor, [])
    os.system('echo 3 > /proc/sys/vm/drop_caches')
    time.sleep(10)
    faas_controller = subprocess.Popen(FAASGRAPH_CONTROLLER_PATH, stdout=subprocess.PIPE)
    local_manager = subprocess.Popen(LOCAL_MANAGER_PATH, stdout=subprocess.PIPE)
    time.sleep(10)
    if app == 'sssp':
        ret = subprocess.run([
            'python3', 
            APPLICAITON_PATH.format(app), 
            dataset + '-weighted'
        ], stdout=subprocess.PIPE)
    elif app == 'cc':
        print("cc_in")
        ret = subprocess.run([
            'python3', 
            APPLICAITON_PATH.format(app), 
            dataset + '-undirected'
        ], stdout=subprocess.PIPE)
        print("cc_out")
    else:
        ret = subprocess.run([
            'python3', 
            APPLICAITON_PATH.format(app), 
            dataset + '-unweighted'
        ], stdout=subprocess.PIPE)
    result_str: List[str] = ret.stdout.decode('utf-8', 'ignore').splitlines()
    print(result_str)
    for s in result_str:
        if s.find('query') != -1:
            query_time = float(re.findall('\d+\.\d+', s)[0])
        elif s.find('store') != -1:
            store_time = float(re.findall('\d+\.\d+', s)[0])
        elif s.find('io') != -1:
            io_time = float(re.findall('\d+\.\d+', s)[0])
        elif s.find('preprocess') != -1:
            preprocess_time = float(re.findall('\d+\.\d+', s)[0])
        elif s.find('latency') != -1:
            end_to_end_time = float(re.findall('\d+\.\d+', s)[0])
        elif s.find('startup') != -1:
            startup_time = float(re.findall('\d+\.\d+', s)[0])
    faas_controller.kill()
    local_manager.kill()
    time.sleep(2)
    local_manager = subprocess.Popen(LOCAL_MANAGER_PATH, stdout=subprocess.PIPE)
    time.sleep(10)
    local_manager.kill()
    time.sleep(2)
    mem_end_flag = True
    gevent.wait([mem_job])
    return end_to_end_time, startup_time, io_time, preprocess_time, query_time, store_time, max_mem - basic_mem

if __name__ == '__main__':
    data = pandas.DataFrame(columns=['dataset', 'framework', 'app', 'end_to_end', 'startup', 'io', 'preprocess', 'query', 'store', 'mem_usage'])
    # for dataset in ['amazon', 'livejournal', 'twitter', 'friendster', 'rmat27']:
    for dataset in ['amazon', 'livejournal', 'twitter']:
        print('------FaaSGraph on {}------'.format(dataset))
        for app in ['bfs', 'cc', 'pr', 'sssp']:
            end_to_end, startup, IO, preprocess, query, store, mem_usage = [], [], [], [], [], [], []
            for i in range(6):
                print('------run {} {}------'.format(app, str(i)))
                e2e, ss, io, pre, q, st, mem = run_faasgraph(dataset, app)
                print('end-to-end: {}s'.format(e2e))
                print('query: {}s'.format(q))
                print('mem usage: {}GB'.format(mem))
                end_to_end.append(e2e)
                startup.append(ss)
                IO.append(io)
                preprocess.append(pre)
                query.append(q)
                store.append(st)
                mem_usage.append(mem)
                data = pandas.concat([data, pandas.DataFrame({'dataset': [dataset], 'framework': ['FaaSGraph'], 'app': [app], 'end_to_end': [e2e], 'startup': [ss], 'io': [io], 'preprocess': [pre], 'query': [q], 'store': [st], 'mem_usage': [mem]})], ignore_index=True)
            data = pandas.concat([data, pandas.DataFrame({'dataset': dataset, 'framework': 'FaaSGraph', 'app': app, 'end_to_end': numpy.average(end_to_end[1:]), 'startup': numpy.average(startup[1:]), 'io': numpy.average(IO[1:]), 'preprocess': numpy.average(preprocess[1:]), 'query': numpy.average(query[1:]), 'store': numpy.average(store[1:]), 'mem_usage': numpy.average(mem_usage[1:])}, index=[0])], ignore_index=True)
            data.to_csv('result.csv')