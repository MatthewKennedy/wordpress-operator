# Copy the wordpress-operator binary into a thin image
# Pinned to the distroless static-debian12 nonroot variant
FROM gcr.io/distroless/static-debian12@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1

COPY rootfs /
ENTRYPOINT ["/wordpress-operator"]
CMD ["help"]
