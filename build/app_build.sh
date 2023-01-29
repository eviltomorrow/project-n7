#!/bin/bash

root_dir=$(pwd)
app_dir=${root_dir}/app
bin_dir=${root_dir}/bin
for name in $(ls ${app_dir}); do
    echo -e "\033[32m=> Building binary(${name})...\033[0m"
    mkdir -p ${bin_dir}/${name}/etc
    cp -rp ${app_dir}/${name}/etc ${bin_dir}/${name}
    echo "go build -ldflags ${2} -gcflags ${4}  -o ${bin_dir}/${name}/bin/${name} ${app_dir}/${name}/main.go"
    go build -ldflags "${2}" -gcflags "${4}"  -o ${bin_dir}/${name}/bin/${name} ${app_dir}/${name}/main.go
    echo -e "\033[32m=> Build Success\033[0m"
done