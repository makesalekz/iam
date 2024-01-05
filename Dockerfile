FROM arigaio/atlas:latest-alpine as atlas
FROM golang:latest AS builder

COPY . /src
WORKDIR /src

RUN git config --global url.https://gitlab-ci-token:glpat-PqK_7yeMpxdsH2NtGssz@gitlab.calendaria.team.insteadOf https://gitlab.calendaria.team && \
    export GOPRIVATE=gitlab.calendaria.team

RUN make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates  \
    netbase \
    && rm -rf /var/lib/apt/lists/ \
    && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app
COPY --from=builder /src/configs/config.yaml /app/

# migration
COPY --from=atlas /atlas /atlas
RUN chmod +x /atlas
COPY --from=builder /src/ent/migrate/migrations/ /migrations/

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
