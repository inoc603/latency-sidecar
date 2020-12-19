# latency-sidecar

## 题目

通过 sidecar，提供 HTTP 接口修改 k8s pod 的网络延时。

## 解决过程记录

### 原理验证

经过简单的搜索，了解到可以使用 [tc](https://man7.org/linux/man-pages/man8/tc.8.html)
等工具设置网络延时。通过 [Dockerfile.poc](./Dockerfile.poc) 中，我们创建一个有相关依赖
的镜像，在本地使用 docker 对原理进行验证：

```
docker build -t latency-test -f Dockerfile.poc
```

运行一个应用容器：

```
docker run -it --rm --name app latency-test bash
```

拿到应用容器 IP:

```
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' app
```

新起容器 ping 应用容器，查看延时：

```
# docker run -it --rm latency-test ping -c 3 172.17.0.2
PING 172.17.0.2 (172.17.0.2) 56(84) bytes of data.
64 bytes from 172.17.0.2: icmp_seq=1 ttl=64 time=0.109 ms
64 bytes from 172.17.0.2: icmp_seq=2 ttl=64 time=0.088 ms
64 bytes from 172.17.0.2: icmp_seq=3 ttl=64 time=0.087 ms

--- 172.17.0.2 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2059ms
rtt min/avg/max/mdev = 0.087/0.094/0.109/0.010 ms
```

另起容器，与应用容器共享网络，使用 tc 设置 20ms 延时：

```
docker run -it --rm --network container:app latency-test tc qdisc add dev eth0 root netem delay 20ms
```

验证延时是否生效：

```
# docker run -it --rm latency-test ping -c 5 172.17.0.2
PING 172.17.0.2 (172.17.0.2) 56(84) bytes of data.
64 bytes from 172.17.0.2: icmp_seq=1 ttl=64 time=44.0 ms
64 bytes from 172.17.0.2: icmp_seq=2 ttl=64 time=20.2 ms
64 bytes from 172.17.0.2: icmp_seq=3 ttl=64 time=20.6 ms
64 bytes from 172.17.0.2: icmp_seq=4 ttl=64 time=22.3 ms
64 bytes from 172.17.0.2: icmp_seq=5 ttl=64 time=20.3 ms

--- 172.17.0.2 ping statistics ---
5 packets transmitted, 5 received, 0% packet loss, time 4018ms
rtt min/avg/max/mdev = 20.246/25.485/43.983/9.277 ms
```

可以发现，除了第一次延时较高外，基本做到了延时增加 20ms。证明此方法可行。
