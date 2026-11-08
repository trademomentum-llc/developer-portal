FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/hello-m2 .

FROM alpine:3.24
COPY --from=build /out/hello-m2 /hello-m2
USER 65532:65532
ENTRYPOINT ["/hello-m2"]
