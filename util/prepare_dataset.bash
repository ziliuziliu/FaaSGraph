mkdir $HOME/data

echo '======build amazon======'
cd $HOME/data
wget https://snap.stanford.edu/data/bigdata/communities/com-amazon.ungraph.txt.gz
gzip -d com-amazon.ungraph.txt.gz
cd ../FaaSGraph/util
go run hash.go $HOME/data/com-amazon.ungraph.txt $HOME/data/amazon.txt TAB 334863
go run txt2csr.go $HOME/data/amazon.txt $HOME/data/amazon-unweighted UNWEIGHTED 334863
go run split.go $HOME/data/amazon.txt $HOME/data/amazon-unweighted UNWEIGHTED 334863 2
go run txt2csr.go $HOME/data/amazon.txt $HOME/data/amazon-weighted WEIGHTED 334863
go run split.go $HOME/data/amazon.txt $HOME/data/amazon-weighted WEIGHTED 334863 2
cd $HOME/data
rm -rf amazon.txt
rm -rf com-amazon.ungraph.txt

echo '======build livejournal======'
cd $HOME/data
wget https://snap.stanford.edu/data/soc-LiveJournal1.txt.gz
gzip -d soc-LiveJournal1.txt.gz
cd ../FaaSGraph/util
go run hash.go $HOME/data/soc-LiveJournal1.txt $HOME/data/livejournal.txt TAB 4847571
go run txt2csr.go $HOME/data/livejournal.txt $HOME/data/livejournal-unweighted UNWEIGHTED DIRECTED 4847571
go run split.go $HOME/data/livejournal.txt $HOME/data/livejournal-unweighted UNWEIGHTED DIRECTED 4847571 2
go run txt2csr.go $HOME/data/livejournal.txt $HOME/data/livejournal-weighted WEIGHTED DIRECTED 4847571
go run split.go $HOME/data/livejournal.txt $HOME/data/livejournal-weighted WEIGHTED DIRECTED 4847571 2
go run txt2csr.go $HOME/data/livejournal.txt $HOME/data/livejournal-undirected UNWEIGHTED UNDIRECTED 4847571
go run split.go $HOME/data/livejournal.txt $HOME/data/livejournal-undirected UNWEIGHTED UNDIRECTED 4847571 2
cd $HOME/data
rm -rf livejournal.txt
rm -rf soc-LiveJournal1.txt

echo '======build twitter======'
cd $HOME/data
wget https://snap.stanford.edu/data/twitter-2010.txt.gz
gzip -d twitter-2010.txt.gz
cd ../FaaSGraph/util
go run hash.go $HOME/data/twitter-2010.txt $HOME/data/twitter.txt SPACE 65608366
go run txt2csr.go $HOME/data/twitter.txt $HOME/data/twitter-unweighted UNWEIGHTED DIRECTED 41652230
go run split.go $HOME/data/twitter.txt $HOME/data/twitter-unweighted UNWEIGHTED DIRECTED 41652230 12
go run txt2csr.go $HOME/data/twitter.txt $HOME/data/twitter-weighted WEIGHTED DIRECTED 41652230
go run split.go $HOME/data/twitter.txt $HOME/data/twitter-weighted WEIGHTED DIRECTED 41652230 24
go run txt2csr.go $HOME/data/twitter.txt $HOME/data/twitter-undirected UNWEIGHTED UNDIRECTED 41652230
go run split.go $HOME/data/twitter.txt $HOME/data/twitter-undirected UNWEIGHTED UNDIRECTED 41652230 24
cd $HOME/data
rm -rf twitter-2010.txt

# echo '======build friendster======'
# cd $HOME/data
# wget https://snap.stanford.edu/data/bigdata/communities/com-friendster.ungraph.txt.gz
# gzip -d com-friendster.ungraph.txt.gz
# cd ../FaaSGraph/util
# go run hash.go $HOME/data/com-friendster.ungraph.txt $HOME/data/friendster.txt TAB 65608366
# go run txt2csr.go $HOME/data/friendster.txt $HOME/data/friendster-unweighted UNWEIGHTED DIRECTED 65608366
# go run split.go $HOME/data/friendster.txt $HOME/data/friendster-unweighted UNWEIGHTED DIRECTED 65608366 16
# go run txt2csr.go $HOME/data/friendster.txt $HOME/data/friendster-weighted WEIGHTED DIRECTED 65608366
# go run split.go $HOME/data/friendster.txt $HOME/data/friendster-weighted WEIGHTED DIRECTED 65608366 28
# go run txt2csr.go $HOME/data/friendster.txt $HOME/data/friendster-undirected UNWEIGHTED UNDIRECTED 65608366
# go run split.go $HOME/data/friendster.txt $HOME/data/friendster-undirected UNWEIGHTED UNDIRECTED 65608366 28
# cd $HOME/data
# rm -rf friendster.txt
# rm -rf com-friendster.ungraph.txt

# echo '======build rmat27======'
# cd $HOME/FaaSGraph/util
# python3 generate_rmat27.py
# go run txt2csr.go $HOME/data/rmat27.txt $HOME/data/rmat27-unweighted UNWEIGHTED DIRECTED 134217728
# go run split.go $HOME/data/rmat27.txt $HOME/data/rmat27-unweighted UNWEIGHTED DIRECTED 134217728 18
# go run txt2csr.go $HOME/data/rmat27.txt $HOME/data/rmat27-weighted WEIGHTED DIRECTED 134217728
# go run split.go $HOME/data/rmat27.txt $HOME/data/rmat27-weighted WEIGHTED DIRECTED 134217728 34
# go run txt2csr.go $HOME/data/rmat27.txt $HOME/data/rmat27-undirected UNWEIGHTED UNDIRECTED 134217728
# go run split.go $HOME/data/rmat27.txt $HOME/data/rmat27-undirected UNWEIGHTED UNDIRECTED 134217728 34
# cd $HOME/data
# rm -rf rmat27.txt