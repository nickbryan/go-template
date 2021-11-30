FROM golang:1.16 AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY service/go.mod .
COPY service/go.sum .
RUN go mod download
COPY ./service .
RUN go build -o /out/gotemplate .

FROM scratch AS bin
COPY --from=build /out/gotemplate /
CMD ["/gotemplate", "server"]
