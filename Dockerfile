FROM golang:latest AS builder

COPY . /src
WORKDIR /src

ARG GOLANG_BUILD_TOKEN
ARG GOLANG_BUILD_TOKEN_PASSWORD

RUN cd $HOME && echo "machine gitlab.calendaria.team login $GOLANG_BUILD_TOKEN password $GOLANG_BUILD_TOKEN_PASSWORD" >> .netrc

RUN GOPRIVATE=gitlab.calendaria.team make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./iam", "-conf", "/data/conf/config.yaml"]
