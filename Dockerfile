FROM gcr.io/distroless/static:nonroot
ARG TARGETPLATFORM

COPY $TARGETPLATFORM/mithras /mithras

EXPOSE 8080

ENTRYPOINT ["/mithras"]
