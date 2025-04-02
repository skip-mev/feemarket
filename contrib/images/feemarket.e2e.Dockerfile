FROM golang:1.24-bullseye AS builder

WORKDIR /src/feemarket
COPY . .

RUN make tidy
RUN make build-test-app

## Prepare the final clear binary
## This will expose the tendermint and cosmos ports alongside
## starting up the sim app and the feemarket daemon
FROM ubuntu:rolling
EXPOSE 26656 26657 1317 9090 7171 26655
ENTRYPOINT ["feemarketd", "start"]

COPY --from=builder /src/feemarket/build/* /usr/local/bin/
RUN apt-get update && apt-get install ca-certificates -y