#!/usr/bin/expect
set user root
set host 47.94.223.143
set password $env(CENTOS_PWD)
set src_path app
set dest_path /root/go/bin
set timeout 30

# build on centos.x86_64
spawn env GOOS=linux GOARCH=amd64 go build ./cmd/app 
expect eof

spawn scp $src_path $user@$host:$dest_path
expect {
    "*password:" { send "$password\n" }
}
interact

puts "\ndeploy success!"
