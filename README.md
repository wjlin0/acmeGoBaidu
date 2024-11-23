<h4 align="center">acmeGoBaidu 是一个用Go编写的自动申请SSL证书并同步到百度CDN的工具。</h4>

<p align="center">
<img src="https://img.shields.io/github/go-mod/go-version/wjlin0/acmeGoBaidu?filename=go.mod" alt="">
<a href="https://github.com/wjlin0/acmeGoBaidu/releases/"><img src="https://img.shields.io/github/release/wjlin0/acmeGoBaidu" alt=""></a> 
<a href="https://github.com/wjlin0/acmeGoBaidu" ><img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/wjlin0/acmeGoBaidu"></a>
<a href="https://github.com/wjlin0/acmeGoBaidu/releases"><img src="https://img.shields.io/github/downloads/wjlin0/acmeGoBaidu/total" alt=""></a> 
<a href="https://github.com/wjlin0/acmeGoBaidu"><img src="https://img.shields.io/github/last-commit/wjlin0/PathScan" alt=""></a> 
<a href="https://www.wjlin0.com/"><img src="https://img.shields.io/badge/wjlin0-blog-green" alt=""></a>
</p>

# Why?

**1. 为什么会选择百度的CDN？**

- 百度CDN支持单栈IPV6回源，且适合动态DDNS域名。
- 百度CDN支持自定义端口回源，适合自定义服务端口。

**2. 域名不备案可以选择百度CDN吗？**
- 不能，百度CDN需要备案的域名。~~如果你的域名没有备案，那么请你选择其他CDN服务商。~~


**3. 为什么要做这个工具？**
- 百度CDN免费的证书只能申请20次（包括续签），是的你没听错。
- 申请证书需要人工操作，且需要等待，这个工具可以自动申请证书并同步到百度CDN。
- ~~因为我喜欢折腾，无意间发现家用IPV6有公网地址，但是是动态的，所以就做了DDNS，但是运营商封了80、443端口，找了一圈，发现只有百度CDN支持单栈IPV6回源、自定义端口回源，所以就有了这个工具。~~

**4. 为什么只支持cloudflare？**
- 只是在自动同步`CNAME`记录，只支持`cloudflare`，在域名解析的时候，可以使用其他服务商，用的是acme的三方包，所以支持大部分市面上的DNS服务商，详情查看[acme.sh](https://github.com/acmesh-official/acme.sh/wiki/dnsapi)。
- 本人使用的是`cloudflare`，如果有其他服务商的需求可以提issue。

**5.推荐什么方式运行？**
- 推荐使用k8s、k3s、docker等容器化方式运行.
- 当然由于工具自带定时任务，也可以直接运行在服务器上。

**6. 如何知道服务商支持的域名解析？**
- 请查看[acme.sh](https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_cf), 并且能够在`acme.sh`中找到对应的服务商的环境变量。

**7. 为什么我无法运行？**
- 你没有`config.yaml`文件，或者没有配置`CLOUDFLARE_EMAIL`、`CLOUDFLARE_API_KEY`、`BAIDUYUN_ACCESSKEY`、`BAIDUYUN_SECRETKEY`这些必须要的环境变量。

# 安装acmeGoBaidu

acmeGoBaidu需要**go1.21**才能安装成功。执行一下命令

```sh
go install -v github.com/wjlin0/acmeGoBaidu/cmd/acmeGoBaidu@latest
```
下载准备运行的[二进制文件](https://github.com/wjlin0/acmeGoBaidu/releases/latest)

- [macOS-arm64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v1.0.2/acmeGoBaidu_1.0.2_macOS_arm64.zip)

- [macOS-amd64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v1.0.2/acmeGoBaidu_1.0.2_macOS_amd64.zip)

- [linux-amd64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v1.0.2/acmeGoBaidu_1.0.2_linux_amd64.zip)

- [windows-amd64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v1.0.2/acmeGoBaidu_1.0.2_windows_amd64.zip)

- [windows-386](https://github.com/wjlin0/acmeGoBaidu/releases/download/v1.0.2/acmeGoBaidu_1.0.2_windows_386.zip)


# 用法

## 命令行
在运行的当前目录下创建一下`config.yaml` 文件
```yaml
acme:
  email: "xxx@xx.com" # acme需要的邮箱
domains:
  - domain: "www.wjlin0.com" # 申请的域名
    provider: "cloudflare" # 域名的服务商
    ipv6: true # 是否开启ipv6
    http2: true # 是否开启http2
    quic: true # 是否开启quic
    baidu: # 百度配置CDN的配置 
      origin: # 来源
        - peer: "https://example.wjlin0.com:443" # 源站地址 https 回源
          host: "www.wjlin0.com" # 回源的域名
          isp: "cm" # 回源的运营商
        - peer: "http://example.wjlin0.com:80" # 源站地址 http 回源
          host: "www.wjlin0.com" # 回源的域名
          isp: "cm" # 回源的运营商
      cname: 
        enabled: true # enable 为 true 时此项有效。是否使用 cname 开启此功能会同步 cname 到 域名服务商 （目前支持 cloudflare，主动提需求才会写其他服务商的代码）
        value: "www.wjlin0.com.a.bdydns.com." # 当然也可以不填写，程序自动补充 默认值是 域名.a.bdydns.com.
      form: dynamic # 默认为"default"，其他可选value："image"表示图片小文件，"download"表示大文件下载，"media"表示流媒体点播，"dynamic"表示动静态加速
      dsa: # 动态加速规则，enabled 为 true 时此项有效。https://cloud.baidu.com/doc/CDN/s/gjwvyex4o#%E8%AF%B7%E6%B1%82%E4%BD%93
        enabled: true
        rules:
          - type: "method"
            value: 'GET;POST;PUT;DELETE;OPTIONS'
```
配置服务商、百度的认证信息
例如：`cloudflare`的认证信息

```shell
export CLOUDFLARE_EMAIL="xxx@xx.com"
export CLOUDFLARE_API_KEY="xxxxx"
export BAIDUYUN_ACCESSKEY="xxxxx"
export BAIDUYUN_SECRETKEY="xxxxx"
```

创建完成后执行以下命令
```sh
acmeGoBaidu
```

## docker
```shell
docker run -d --name acmeGoBaidu -e CLOUDFLARE_EMAIL="xxx@xx.com" -e CLOUDFLARE_API_KEY="xxxxx" -e BAIDUYUN_ACCESSKEY="xxxx" -e BAIDUYUN_ACCESSKEY="xxxx" -e CRON="0 0 * * 1" -v ./data/config.yaml:/app/config/config.yaml -v ./data/certs:/app/certs wjlin0/acmegobaidu:latest
```

## K8s
编写一下`yaml`文件

```yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: acme-go-baidu-config
data:
  config.yaml: |
    acme:
      email: "xxx@xx.com" # acme需要的邮箱
    domains:
      - domain: "www.wjlin0.com" # 申请的域名
        provider: "cloudflare" # 域名的服务商
        ipv6: true # 是否开启ipv6
        http2: true # 是否开启http2
        quic: true # 是否开启quic
        baidu: # 百度配置CDN的配置 
          origin: # 来源
            - peer: "https://example.wjlin0.com:443" # 源站地址 https 回源
              host: "www.wjlin0.com" # 回源的域名
              isp: "cm" # 回源的运营商
            - peer: "http://example.wjlin0.com:80" # 源站地址 http 回源
              host: "www.wjlin0.com" # 回源的域名
              isp: "cm" # 回源的运营商
          cname:
            enabled: true # enable 为 true 时此项有效。是否使用 cname 开启此功能会同步 cname 到 域名服务商 （目前支持 cloudflare，主动提需求才会写其他服务商的代码）
            value: "www.wjlin0.com.a.bdydns.com." # 当然也可以不填写，程序自动补充 默认值是 域名.a.bdydns.com.
          form: dynamic # 默认为"default"，其他可选value："image"表示图片小文件，"download"表示大文件下载，"media"表示流媒体点播，"dynamic"表示动静态加速
          dsa: # 动态加速规则，enabled 为 true 时此项有效。https://cloud.baidu.com/doc/CDN/s/gjwvyex4o#%E8%AF%B7%E6%B1%82%E4%BD%93
            enabled: true
            rules:
              - type: "method"
                value: 'GET;POST;PUT;DELETE;OPTIONS'
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: acme-go-baidu-deployment
spec:
  selector:
    matchLabels:
      app: acme-go-baidu
  template:
    metadata:
      labels:
        app: acme-go-baidu
    spec:
      restartPolicy: Always
      containers:
        - name: acme-go-baidu-container
          image: wjlin0/acmegobaidu:latest
          resources:
            limits:
              memory: '32Mi'
              cpu: '50m'
          env:
            - name: CONFIG_PATH
              value: '/app/config/config.yaml'
            - name: JSON_PATH
              value: '/app/certs/certificate.json'
            - name: CRON
              value: '0 0 * * 1' # 每周一凌晨执行
            - name: CLOUDFLARE_EMAIL
              value: 'xxx@xxx.com'
            - name: CLOUDFLARE_API_KEY
              value: 'xxxxx'
            - name: BAIDUYUN_ACCESSKEY
              value: 'xxxxx'
            - name: BAIDUYUN_SECRETKEY
              value: 'xxxxx'
          command:
            - "acmeGoBaidu"
          volumeMounts:
            - mountPath: '/app'
              subPath: 'acmeGoBaidu.yaml'
              name: config
              readOnly: true
            - mountPath: '/app/certs'
              name: certs 
      volumes:
        - name: config
          configMap:
            name: acme-go-baidu-config
        - name: certs
          persistentVolumeClaim:
            claimName: acme-go-baidu-pvc
```
创建pvc

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: acme-go-baidu-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 100Mi
```


```shell
# 创建 pvc
kubectl create -f pvc.yaml
# 运行
kubectl apply -f acmeGoBaidu.yaml
```


# 注意

**根域名可能无法配置`CNAME` 记录，配置后，不生效**
