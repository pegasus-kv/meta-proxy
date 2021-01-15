FROM cr.d.xiaomi.net/container/xiaomi_centos7:release

WORKDIR /home/work

COPY bin/meta-proxy .
COPY config/yaml/*.yml .
