FROM graph_base
WORKDIR /lambda_executor
RUN mkdir /lambda_executor/src/graphcalc
COPY graph_calc.go /lambda_executor/src/graphcalc
RUN go build -o proxy ./src
CMD ./start_proxy.sh
