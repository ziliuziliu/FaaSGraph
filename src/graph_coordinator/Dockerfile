FROM golang:1.18-alpine
LABEL label="graph"
USER root
WORKDIR /graph_coordinator
COPY . .
EXPOSE 5000
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN go build -o graph_coordinator ./src
CMD ./graph_coordinator
