FROM golang:latest AS builder

COPY . /src
WORKDIR /src

RUN git config --global url."https://gitlab-ci-token:glpat-ZfWMnTP-NvZYJPCxfvzB@gitlab.calendaria.team/".insteadOf https://gitlab.calendaria.team/

RUN --mount=type=ssh GOPRIVATE=gitlab.calendaria.team make build

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
