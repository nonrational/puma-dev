#!/usr/bin/env bash

# https://travis-ci.org/puma/puma-dev/jobs/644550988?utm_medium=notification&utm_source=github_status
# hostname: f64d9378-ccf4-47f5-8ab6-b9ee4adf0567@1.worker-org-85d846cc5-94hp6.gce-production-1
# version: v6.2.6 https://github.com/travis-ci/worker/tree/ba21bd30589fd152126e13df30e0cc78ccdf2837
# instance: travis-job-ab6205eb-3305-47f4-b0a7-ae5da92bfe11 travis-ci-sardonyx-xenial-1553530528-f909ac5 (via amqp)
# startup: 5.95650086s

BUILDID="build-$RANDOM"
# INSTANCE="travisci/ci-sardonyx:packer-1558623664-f909ac5"
INSTANCE="travisci/ci-go:packer-1494864792"

docker run --name $BUILDID -dit $INSTANCE /sbin/init
docker exec -it $BUILDID bash -l

exit 0
#### inside docker

su - travis

cat <<EOF > build.sh
#!/usr/bin/env bash

set -x
set -e

git clone --depth=50 https://github.com/puma/puma-dev.git puma/puma-dev
cd puma/puma-dev
git fetch origin +refs/pull/221/merge:

travis_export_go 1.10.x github.com/puma/puma-dev
travis_setup_go

export GOPATH="/home/travis/gopath"
export PATH="/home/travis/gopath/bin:/home/travis/.gimme/versions/go1.10.8.linux.amd64/bin:/home/travis/bin:/home/travis/bin:/home/travis/.local/bin:/usr/local/lib/jvm/openjdk11/bin:/opt/pyenv/shims:/home/travis/.phpenv/shims:/home/travis/perl5/perlbrew/bin:/home/travis/.nvm/versions/node/v8.12.0/bin:/home/travis/.rvm/gems/ruby-2.5.3/bin:/home/travis/.rvm/gems/ruby-2.5.3@global/bin:/home/travis/.rvm/rubies/ruby-2.5.3/bin:/home/travis/gopath/bin:/home/travis/.gimme/versions/go1.11.1.linux.amd64/bin:/usr/local/maven-3.6.0/bin:/usr/local/cmake-3.12.4/bin:/usr/local/clang-7.0.0/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/home/travis/.rvm/bin:/home/travis/.phpenv/bin:/opt/pyenv/bin:/home/travis/.yarn/bin"
export GO111MODULE="auto"

gimme version
go version
go env
go get -t ./...

EOF
