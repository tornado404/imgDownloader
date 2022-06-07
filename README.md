# version
fileload.exe -version
# download with one coroutine
fileload.exe -f xx.zip http://xxx.com/xx.zip
# download with 10 coroutines
fileload.exe -c 10 -f xx.zip http://xxx.com/xx.zip
# download with coroutines and specify the size of chunk to 1M
fileload.exe -c 10 -s 1000000 -f xx.zip http://xxx.com/xx.zip
# only sum the hash
fileload.exe -v xx.zip