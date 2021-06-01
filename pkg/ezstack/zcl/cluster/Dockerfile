# for using generate.js. Usage:
#
# 	$ docker build -t zcl-generator .
# 	$ docker run --rm zcl-generator > clustersandattributes.gen.go

FROM node

WORKDIR /workspace

CMD ["/workspace/generate.js"]

RUN npm install zigbee-herdsman

ADD generate.js /workspace/
