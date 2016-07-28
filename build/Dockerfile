FROM centos:7

RUN printf '[mf2]\n\
name=GrayMeta Metafarm Repository - staging\n\
baseurl=https://gmbuild1.graymeta.com/el/7/mf2/staging/$basearch/\n\
enabled=1\n\
gpgcheck=0\n\
metadata_expire=0\n\
sslverify=False' \
>> /etc/yum.repos.d/mf2.repo

RUN yum -y install \
    git \
    make \
    wget \
    gm-devtools \
    && yum clean all

ENV goversion 1.6.3
ENV gofile go${goversion}.linux-amd64.tar.gz
ENV gourl https://storage.googleapis.com/golang/${gofile}

RUN wget -q -O /usr/local/${gofile} ${gourl} \
  && mkdir /usr/local/go \
  && tar -xzf /usr/local/${gofile} -C /usr/local/go --strip 1

CMD cd /mnt/src/github.com/graymeta/stow && PATH=/usr/local/go/bin:$PATH GOPATH=/mnt make test
