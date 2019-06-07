# Stage 0, "build-stage", based on Node.js, to build and compile the frontend
FROM node:10.8.0 as build-forntend
WORKDIR /app
#COPY package*.json /app/
COPY ./kiali-ui /app/
ENV NODE_OPTIONS=--max_old_space_size=4096
RUN npm install -g yarn
RUN yarn
RUN yarn build
RUN ls -ltr

#COPY --from=build-stage /app/dist/out/ /usr/share/nginx/html
#stage 2, copy forntend dist and go binary
FROM centos:centos7

LABEL maintainer="kiali-dev@googlegroups.com"

ENV KIALI_HOME=/opt/kiali \
    PATH=$KIALI_HOME:$PATH
RUN echo $GOPATH
WORKDIR $KIALI_HOME
COPY ./kiali/kiali $KIALI_HOME/

RUN mkdir $KIALI_HOME/console/

COPY --from=build-forntend /app/build/ $KIALI_HOME/console/

RUN chgrp -R 0 $KIALI_HOME/console && \
    chmod -R g=u $KIALI_HOME/console

ENTRYPOINT ["/opt/kiali/kiali"]
