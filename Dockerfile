FROM cr.d.xiaomi.net/container/xiaomi_centos7:release

WORKDIR /home/work

COPY bin/meta-proxy .
COPY meta-proxy.yml .
COPY perfcounter.json .

ENTRYPOINT ["bash", "-c", "./meta-proxy meta-proxy.yml"]
