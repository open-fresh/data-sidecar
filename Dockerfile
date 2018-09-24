FROM scratch
ADD data-sidecar /data-sidecar
ENTRYPOINT ["/data-sidecar"]
