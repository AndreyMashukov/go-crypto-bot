FROM pytorch/manylinux-cpu:latest
# amashukov/golang:1.21.1-pytorch
# python3.11.6

RUN yum update -y
RUN yum -y install libffi-devel zlib-devel libdl.so.2 glibc util-linux libutil.so.1 gcc openssl-devel bzip2-devel --skip-broken
RUN yum -y group install "Development Tools" --skip-broken

# Install OpenSSL
# todo: updated to latest is https://openssl.org/source/openssl-3.3.1.tar.gz
RUN cd /usr/src \
    && wget https://openssl.org/source/openssl-1.1.1q.tar.gz --no-check-certificate \
    && tar -xzvf openssl-1.1.1q.tar.gz \
    && cd openssl-1.1.1q \
    && ./config --prefix=/usr --openssldir=/etc/ssl --libdir=lib no-shared zlib-dynamic \
    && make && make install \
    && openssl version

# Install Python 3.11.6
RUN wget https://www.python.org/ftp/python/3.11.6/Python-3.11.6.tgz \
    && tar xzf Python-3.11.6.tgz \
    && cd Python-3.11.6 \
    && ./configure --prefix=/usr/local/ --with-openssl=/usr --enable-optimizations  \
    && make && make altinstall && cd .. && rm -rf Python-3.11.6

RUN rm /usr/bin/python3
RUN ln -s /usr/local/bin/python3.11 /usr/bin/python3
RUN ln -s /usr/local/bin/pip3.11 /usr/bin/pip
RUN rm /usr/bin/pip3
RUN ln -s /usr/local/bin/pip3.11 /usr/bin/pip3

RUN pip install setuptools --force-reinstall
RUN pip install wheel --force-reinstall
RUN pip uninstall comtypes
RUN pip install --no-cache-dir comtypes

RUN yum -y install python3-devel
RUN yum -y install --skip-broken curl gcc-c++ libaio-devel git mc make \
    && yum clean all \
    && rm -r -f /var/cache/yum/*

RUN pip3.11 install --upgrade pip
RUN pip3.11 install --trusted-host pypi.org setuptools-rust transformers torch

# Install Golang 1.21.1
RUN wget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=$HOME/go
ENV PATH=$PATH:$GOPATH/bin

CMD ["tail", "-f", "/dev/null"]