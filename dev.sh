go build .

# ./mydocker run -ti -v /root/gos_open/mydocker/test_volume:/root/aa --name qo sh

./mydocker run -d --name container1 -v /root/gos_open/mydocker/test_volume:/w1 busybox top

./mydocker run -d --name container2 -v /root/gos_open/mydocker/test_volume2:/w2 busybox top