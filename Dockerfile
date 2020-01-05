FROM gcr.io/distroless/base:latest

COPY cmd/check/check /opt/resource/check
COPY cmd/in/in /opt/resource/in
COPY cmd/out/out /opt/resource/out