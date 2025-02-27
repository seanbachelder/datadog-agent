########################################
# Preparation stage: layout and chmods #
########################################

FROM ubuntu:22.04 as builder

WORKDIR /output

COPY datadog-cluster-agent.arm64 opt/datadog-agent/bin/datadog-cluster-agent
COPY ./conf.d etc/datadog-agent/conf.d
COPY ./datadog-cluster.yaml etc/datadog-agent/datadog-cluster.yaml
COPY ./security-agent-policies/compliance/containers/ etc/datadog-agent/compliance.d
COPY ./install_info etc/datadog-agent/install_info
COPY entrypoint.sh .
COPY readsecret.sh readsecret_multiple_providers.sh ./

RUN chmod 755 entrypoint.sh \
    && chown root:root readsecret.sh readsecret_multiple_providers.sh \
    && chmod 550 readsecret.sh readsecret_multiple_providers.sh\
    && chmod g+r,g+w,g+X -R etc/datadog-agent/ \
    && chmod +x opt/datadog-agent/bin/datadog-cluster-agent \
    && ln -s /opt/datadog-agent/bin/datadog-cluster-agent opt/datadog-agent/bin/agent

FROM builder AS nosys-seccomp
COPY nosys-seccomp/nosys.c   /tmp/nosys.c
COPY nosys-seccomp/nosys.sym /tmp/nosys.sym
ENV DEBIAN_FRONTEND=noninteractive
RUN apt update && apt install --no-install-recommends -y gcc libc6-dev libseccomp-dev
RUN gcc -pipe -Wall -Wextra -O2 -shared -fPIC -Wl,--version-script=/tmp/nosys.sym -o /tmp/nosys.so /tmp/nosys.c -lseccomp

####################################
# Actual docker image construction #
####################################

FROM ubuntu:22.04

LABEL maintainer "Datadog <package@datadoghq.com>"

ARG CIBUILD
# NOTE about APT mirrorlists:
# It seems that this feature could use some improvement. If you just get mirrorlist
# from mirrors.ubuntu.com/mirrors.txt, it might contain faulty mirrors that either
# cause `apt update` to fail with exit code 100 or make it hang on `0% [Working]`
# indefinitely. Therefore we create a mirrorlist with the 2 mirrors that we know
# should be reliable enough in combination and also well maintained.
 RUN if [ "$CIBUILD" = "true" ]; then \
  echo "http://us-east-1.ec2.ports.ubuntu.com/ubuntu-ports\tpriority:1\nhttp://ports.ubuntu.com/ubuntu-ports" > /etc/apt/mirrorlist && \
  sed -i -e 's#http://ports.ubuntu.com\S*#mirror+file:/etc/apt/mirrorlist#g' /etc/apt/sources.list; \
   fi

ENV PATH="/opt/datadog-agent/bin/:$PATH" \
    DOCKER_DD_AGENT="true" \
    # Allow User Group to exec the secret backend script.
    DD_SECRET_BACKEND_COMMAND_ALLOW_GROUP_EXEC_PERM="true"

RUN apt-get update \
    && apt full-upgrade -y \
    && apt-get install --no-install-recommends -y ca-certificates curl libseccomp2 tzdata \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /output /

# Allow running as an unprivileged user:
# - General case is the dd-agent user
# - OpenShift uses a random UID in the root group
#
# Containerd does not preserve permissions when mounting a volume on top
# of an empty folder. Creating .placeholder files as a workaround.
#
RUN adduser --system --no-create-home --disabled-password --ingroup root dd-agent \
    && mkdir -p /var/log/datadog/ /conf.d \
    && touch /var/log/datadog/.placeholder \
    && touch /tmp/.placeholder \
    && chown -R dd-agent:root /etc/datadog-agent/ /var/log/datadog/ /conf.d /tmp/ \
    && chmod g+r,g+w,g+X -R /etc/datadog-agent/ /var/log/datadog/ /conf.d /tmp/

# Ensure the glibc doesn't try to call syscalls that may not be supported
COPY --from=nosys-seccomp /tmp/nosys.so /lib/x86_64-linux-gnu/nosys.so
ENV LD_PRELOAD=/lib/x86_64-linux-gnu/nosys.so

# Incompatible with the custom metrics API on port 443
# Set DD_EXTERNAL_METRICS_PROVIDER_PORT to a higher value to run as non-root
# USER dd-agent

# Leave following directories RW to allow use of readonly rootfs
VOLUME ["/etc/datadog-agent", "/var/log/datadog", "/tmp"]

ENTRYPOINT ["/entrypoint.sh"]

# No docker healthcheck, use a HTTP check
# on port 5005 and/or 443 on Kubernetes

CMD ["datadog-cluster-agent", "start"]
