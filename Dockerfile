FROM golang:1.22-alpine AS build

RUN adduser --uid 10000 --disabled-password dce

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/divisas_cuba_exporter /divisas_cuba_exporter
USER dce
CMD ["/divisas_cuba_exporter"]
