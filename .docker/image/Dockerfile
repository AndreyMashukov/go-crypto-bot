FROM golang:1.21.1-alpine
# amashukov/golang:1.21.1-ml

RUN apk add python3=~3.11.6 python3-dev gcc libc-dev musl-dev pkgconfig g++
RUN apk add py3-scikit-learn py3-pandas py3-joblib
RUN apk add py3-numpy
RUN apk cache clean
