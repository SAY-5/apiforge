FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go test ./... && go build -o /apiforge ./cmd/apiforge

FROM alpine:3.20
RUN adduser -D -u 1001 af
USER af
COPY --from=build /apiforge /usr/local/bin/apiforge
ENTRYPOINT ["apiforge"]
