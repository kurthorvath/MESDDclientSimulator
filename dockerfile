FROM alnoda/alnoda-workspace:latest
EXPOSE 8020
CMD  mkdir /APP
COPY ./main /APP/
WORKDIR /APP
CMD chmod 777 main
CMD ./main