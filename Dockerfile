# FROM quay.io/prometheus/busybox:latest
FROM s390x/busybox:latest
COPY ds8k.yaml /etc/ds8k-exporter/ds8k.yaml
COPY ds8k-exporter /bin/ds8k-exporter
EXPOSE 9710
ENTRYPOINT ["/bin/ds8k-exporter"]
CMD ["--config.file=/etc/ds8k-exporter/ds8k.yaml"]