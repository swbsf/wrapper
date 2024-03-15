FROM alpine:latest
RUN wget https://github.com/loft-sh/vcluster/releases/download/v0.19.4/vcluster-linux-arm64
RUN chmod +x vcluster-linux-arm64
RUN mv vcluster-linux-arm64 /usr/local/bin/vcluster
RUN apk add kubectl curl bash pipx git
RUN pipx install copier
ENV PATH="${PATH}:/root/.local/bin"

