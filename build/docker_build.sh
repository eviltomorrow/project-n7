#!/bin/bash

root_dir=$(pwd)
app_dir=${root_dir}/app
for name in $(ls ${app_dir}); do
    docker build --target prod -t ${name} . --build-arg APPNAME=${name} --build-arg MAINVERSION=${1} --build-arg GITSHA=${2} --build-arg BUILDTIME=${3}
    echo -e "\033[32m=> Build Success\033[0m"
done