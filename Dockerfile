FROM openeuler/openeuler:23.03 as BUILDER
RUN dnf update -y && \
    dnf install -y golang git make && \
    go env -w GOPROXY=https://goproxy.cn,direct

# build binary
WORKDIR /go/src/github.com/opensourceways/software-package-gateway
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o software-package-gateway .

# copy binary config and utils
FROM openeuler/openeuler:22.03
RUN dnf -y update && \
    dnf in -y shadow && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 gateway && \
    useradd -u 1000 -g gateway -s /sbin/nologin -m gateway && \
    echo "umask 027" >> /home/gateway/.bashrc && \
    echo 'set +o history' >> /home/gateway/.bashrc && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && rm -rf /tmp/* && \
    mkdir /opt/app -p && chmod 700 /opt/app && chown 1000:1000 /opt/app

USER gateway

WORKDIR /opt/app/

COPY --chown=gateway --from=BUILDER /go/src/github.com/opensourceways/software-package-gateway/software-package-gateway /opt/app/software-package-gateway

RUN chmod 550 /opt/app/software-package-gateway

ENTRYPOINT ["/opt/app/software-package-gateway"]
