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
go run txt2csr.go $HOME/data/livejournal.txt $HOME/data/livejournal-unweighted UNWEIGHTED 4847571
go run split.go $HOME/data/livejournal.txt $HOME/data/livejournal-unweighted UNWEIGHTED 4847571 2
go run txt2csr.go $HOME/data/livejournal.txt $HOME/data/livejournal-weighted WEIGHTED 4847571
go run split.go $HOME/data/livejournal.txt $HOME/data/livejournal-weighted WEIGHTED 4847571 2
cd $HOME/data
rm -rf livejournal.txt
rm -rf soc-LiveJournal1.txt

echo '======build twitter======'
cd $HOME/data
wget https://snap.stanford.edu/data/twitter-2010.txt.gz
gzip -d twitter-2010.txt.gz
cd ../FaaSGraph/util
go run txt2csr.go $HOME/data/twitter-2010.txt $HOME/data/twitter-unweighted UNWEIGHTED 41652230
go run split.go $HOME/data/twitter-2010.txt $HOME/data/twitter-unweighted UNWEIGHTED 41652230 12
go run txt2csr.go $HOME/data/twitter-2010.txt $HOME/data/twitter-weighted WEIGHTED 41652230
go run split.go $HOME/data/twitter-2010.txt $HOME/data/twitter-weighted WEIGHTED 41652230 16
cd $HOME/data
rm -rf twitter-2010.txt

# echo '======build friendster======'
# cd $HOME/data
# wget https://snap.stanford.edu/data/bigdata/communities/com-friendster.ungraph.txt.gz
# gzip -d com-friendster.ungraph.txt.gz
# cd ../FaaSGraph/util
# go run hash.go $HOME/data/com-friendster.ungraph.txt $HOME/data/friendster.txt TAB 65608366
# go run txt2csr.go $HOME/data/friendster.txt $HOME/data/friendster-unweighted UNWEIGHTED 65608366
# go run split.go $HOME/data/friendster.txt $HOME/data/friendster-unweighted UNWEIGHTED 65608366 16
# go run txt2csr.go $HOME/data/friendster.txt $HOME/data/friendster-weighted WEIGHTED 65608366
# go run split.go $HOME/data/friendster.txt $HOME/data/friendster-weighted WEIGHTED 65608366 20
# cd $HOME/data
# rm -rf friendster.txt
# rm -rf com-friendster.ungraph.txt
