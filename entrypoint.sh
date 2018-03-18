#! /usr/bin/env sh

case $2 in
server)
    exec swagger "$@" --template-dir=/templates --exclude-main
    ;;
client)
    exec swagger "$@" --template-dir=/templates
    ;;
*)
    exec swagger "$@"
    ;;
esac
