import networkit as nk

scale = 27
edge_factor = 16
a = 0.57
b = 0.19
c = 0.19
d = 1 - (a + b + c)

rmat_gen = nk.generators.RmatGenerator(scale, edge_factor, a, b, c, d)


graph = rmat_gen.generate()

writer = nk.graphio.EdgeListWriter(separator=' ', firstNode=0)

writer.write(graph, "/home/ubuntu/data/rmat27.txt")

