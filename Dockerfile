FROM ignitehq/cli
USER root
ADD . .
RUN ignite chain build