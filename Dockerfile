# for hadolint Error and warning codes - https://github.com/hadolint/hadolint.wiki.git

# Development image
FROM ajlocalau-docker-prod-public.jfrog.io/golang:1.16.7-alpine3.14 AS BUILD-ENV
SHELL ["/bin/ash", "-eo", "pipefail", "-c"]
ARG GOOS_VAL 
ARG GOARCH_VAL
RUN apk --no-cache add ca-certificates openssl git curl build-base 
ARG cert_location=/usr/local/share/ca-certificates
# Get certificate from "github.com"
RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM>${cert_location}/github.crt
# Get certificate from "proxy.golang.org"
RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM>${cert_location}/proxy.golang.crt
RUN openssl s_client -showcerts -connect storage.googleapis.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM>${cert_location}/storage.googleapis.crt
RUN wget --no-check-certificate -q https://storage.googleapis.com/kubernetes-release/release/`curl -k -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl
# Update certificates
RUN update-ca-certificates

RUN rm -rf /var/cache/apk/*
WORKDIR /app
# Download dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy source
COPY . .
# Build binary
RUN go build -o /go/bin/dalek ./cmd/dalek 

# Production image
FROM ajlocalau-docker-prod-public.jfrog.io/alpine:3.10
# Create Non Privileged user
RUN addgroup --gid 101 dalek && \
    adduser -S --uid 101 --ingroup dalek dalek
# Run as Non Privileged user
USER dalek

COPY --from=BUILD-ENV /go/bin/dalek /go/bin/dalek
COPY --from=BUILD-ENV /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=BUILD-ENV /usr/local/bin/kubectl /usr/local/bin/kubectl

ENTRYPOINT ["/go/bin/dalek"]
