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

## 使用 TC 命令行实现接口

程序中调用命令行工具并不是一个特别好的办法，但我们可以先通过调用 tc
命令来实现，以作为基准，再选择更好的 rtnetlink api。

在 [pods.yml](pods.yml) 中，定义一个简单的 pod，同时启动应用容器和 agent。

启动应用和 agent:

```
make k8s/run
```

测试设置延时：

```
# make k8s/test
If you don't see a command prompt, try pressing enter.
target ip: 10.1.0.20, agent port: 8080
PING 10.1.0.20 (10.1.0.20) 56(84) bytes of data.
64 bytes from 10.1.0.20: icmp_seq=1 ttl=64 time=0.649 ms
64 bytes from 10.1.0.20: icmp_seq=2 ttl=64 time=0.111 ms
64 bytes from 10.1.0.20: icmp_seq=3 ttl=64 time=0.248 ms

--- 10.1.0.20 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2066ms
rtt min/avg/max/mdev = 0.111/0.336/0.649/0.228 ms
set latency to 20ms
latency set to 20ms
PING 10.1.0.20 (10.1.0.20) 56(84) bytes of data.
64 bytes from 10.1.0.20: icmp_seq=1 ttl=64 time=20.3 ms
64 bytes from 10.1.0.20: icmp_seq=2 ttl=64 time=20.5 ms
64 bytes from 10.1.0.20: icmp_seq=3 ttl=64 time=21.4 ms

--- 10.1.0.20 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2004ms
rtt min/avg/max/mdev = 20.261/20.722/21.420/0.501 ms
reset latency
latency set to 0s
ping target
PING 10.1.0.20 (10.1.0.20) 56(84) bytes of data.
64 bytes from 10.1.0.20: icmp_seq=1 ttl=64 time=0.074 ms
64 bytes from 10.1.0.20: icmp_seq=2 ttl=64 time=0.226 ms
64 bytes from 10.1.0.20: icmp_seq=3 ttl=64 time=0.195 ms

--- 10.1.0.20 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2049ms
rtt min/avg/max/mdev = 0.074/0.165/0.226/0.065 ms
pod "client" deleted
```

这里为了简单，没有写程序检查结果，暂时靠肉眼观察输出即可。
