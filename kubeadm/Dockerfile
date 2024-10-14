FROM alpine:3.20 as build

ARG RELEASE="v1.31.1"

RUN apk update && \
    apk upgrade && \
    apk add curl
RUN curl -Lo /usr/local/bin/kubeadm https://dl.k8s.io/release/${RELEASE}/bin/linux/amd64/kubeadm
RUN chmod +x /usr/local/bin/kubeadm

FROM alpine:3.20 
COPY --from=build /usr/local/bin/kubeadm /usr/local/bin/kubeadm