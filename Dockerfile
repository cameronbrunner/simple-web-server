FROM centos:7

WORKDIR /app/
ADD app-linux ./
ADD *.html ./

EXPOSE 8085

ENTRYPOINT [ "/app/app-linux" ]
