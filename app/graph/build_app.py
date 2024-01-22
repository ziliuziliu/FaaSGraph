import os

def build_app(functions) :
    for function in functions:
        print('build image for {}'.format(function))
        os.system('docker build --no-cache -t graph-{} {}'.format(function, function))

if __name__ == '__main__':
    functions = ['bfs', 'cc', 'pr', 'sssp', 'ccpull', 'prpush']
    build_app(functions)
