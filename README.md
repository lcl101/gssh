# gssh
一个自用的ssh登录相应工具，包括：
- gal：记住密码自动登录服务，例如登录阿里服务器：./gal aliserver
- grr：记住密码远程执行密码，例如在阿里服务器执行ls命令： ./grr aliserver 'ls -lart'
- gcp：记住密码，进行服务器文件拷贝，例如从服务器拷贝文件（类似scp）：./gcp aliserver:~/test.pdf ./test.pdf
