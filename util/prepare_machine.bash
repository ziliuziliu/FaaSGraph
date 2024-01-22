cd ../..
echo "==========install golang=========="
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> .bashrc
source .bashrc
rm -rf go1.21.6.linux-amd64.tar.gz
echo "==========install docker=========="
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg lsb-release
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
echo "==========build image=========="
cd FaaSGraph/src/lambda_executor
docker build -t graph_base .
cd ../../app/graph
python3 build_app.py
cd ../../src/graph_coordinator
docker build -t graph-coordinator .
echo "==========install python package=========="
sudo apt install -y python3-pip
sudo pip3 install gevent pandas