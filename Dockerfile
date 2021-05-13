FROM golang:1.16 as builder

# Metadata params
ARG VERSION
ARG BUILD_DATE
ARG VCS_URL
ARG VCS_REF
ARG NAME
ARG VENDOR
ARG GOFLAGS

# Setup environment
WORKDIR /go/src/github.com/$VENDOR/$NAME
ADD . /go/src/github.com/$VENDOR/$NAME

# Test
RUN make test

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app -ldflags "-X main.Version=${VCS_REF} -X main.BuildTime=${BUILD_DATE}" .

# =================================================================

# Start from clean image
FROM scratch
# Metadata params
ARG VERSION
ARG BUILD_DATE
ARG VCS_URL
ARG VCS_REF
ARG NAME
ARG VENDOR
ARG GOFLAGS

# Copy binary from first stage
COPY --from=builder /go/src/github.com/$VENDOR/$NAME/app .

# Metadata
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name=$NAME \
      org.label-schema.description="A Go server for video broadcast via Kafka" \
      org.label-schema.url="https://github.com/dstockton/kafka-video-broadcast" \
      org.label-schema.vcs-url=https://github.com/$VENDOR/$VCS_URL \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vendor=$VENDOR \
      org.label-schema.version=$VERSION \
      org.label-schema.docker.schema-version="1.0" \
      org.label-schema.docker.cmd="docker run -d $VENDOR/$NAME"

# Default command
CMD ["./app"]
