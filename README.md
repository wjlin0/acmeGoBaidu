<h4 align="center">acmeGoBaidu 是一个用Go编写的自动申请SSL证书并同步到百度CDN的工具。</h4>

<p align="center">
<img src="https://img.shields.io/github/go-mod/go-version/wjlin0/acmeGoBaidu?filename=go.mod" alt="">
<a href="https://github.com/wjlin0/acmeGoBaidu/releases/"><img src="https://img.shields.io/github/release/wjlin0/acmeGoBaidu" alt=""></a> 
<a href="https://github.com/wjlin0/acmeGoBaidu" ><img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/wjlin0/acmeGoBaidu"></a>
<a href="https://github.com/wjlin0/acmeGoBaidu/releases"><img src="https://img.shields.io/github/downloads/wjlin0/acmeGoBaidu/total" alt=""></a> 
<a href="https://github.com/wjlin0/acmeGoBaidu"><img src="https://img.shields.io/github/last-commit/wjlin0/PathScan" alt=""></a> 
<a href="https://blog.wjlin0.com/"><img src="https://img.shields.io/badge/wjlin0-blog-green" alt=""></a>
</p>



# 安装acmeGoBaidu

acmeGoBaidu需要**go1.21**才能安装成功。执行一下命令

```sh
go install -v github.com/wjlin0/acmeGoBaidu/cmd/acmeGoBaidu@latest
```
下载准备运行的[二进制文件](https://github.com/wjlin0/acmeGoBaidu/releases/latest)

- [macOS-arm64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v2.1.4/acmeGoBaidu_2.1.4_macOS_arm64.zip)

- [macOS-amd64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v2.1.4/acmeGoBaidu_2.1.4_macOS_amd64.zip)

- [linux-amd64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v2.1.4/acmeGoBaidu_2.1.4_linux_amd64.zip)

- [windows-amd64](https://github.com/wjlin0/acmeGoBaidu/releases/download/v2.1.4/acmeGoBaidu_2.1.4_windows_amd64.zip)

- [windows-386](https://github.com/wjlin0/acmeGoBaidu/releases/download/v2.1.4/acmeGoBaidu_2.1.4_windows_386.zip)


# 用法

## 命令行
在运行的当前目录下创建一下`config.yaml` 文件
```yaml
acme:
  email: "xxx@xx.com" # acme需要的邮箱
domains:
  - domain: "www.wjlin0.com" # 申请的域名
    provider: "cloudflare" # 域名的服务商
    baidu: # 百度配置CDN的配置 
      origin: # 来源
        - peer: "https://example.wjlin0.com:443" # 源站地址 https 回源
          host: "www.wjlin0.com" # 回源的域名
        - peer: "http://example.wjlin0.com:80" # 源站地址 http 回源
          host: "www.wjlin0.com" # 回源的域名
      form: dynamic # 默认为"default"，其他可选value："image"表示图片小文件，"download"表示大文件下载，"media"表示流媒体点播，"dynamic"表示动静态加速
      dsa: # 动态加速规则，enable 为 true 时此项有效。https://cloud.baidu.com/doc/CDN/s/gjwvyex4o#%E8%AF%B7%E6%B1%82%E4%BD%93
        enable: true
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
        baidu: # 百度配置CDN的配置 
          origin: # 来源
            - peer: "https://example.wjlin0.com:443" # 源站地址 https 回源
              host: "www.wjlin0.com" # 回源的域名
            - peer: "http://example.wjlin0.com:80" # 源站地址 http 回源
              host: "www.wjlin0.com" # 回源的域名
          form: dynamic # 默认为"default"，其他可选value："image"表示图片小文件，"download"表示大文件下载，"media"表示流媒体点播，"dynamic"表示动静态加速
          dsa: # 动态加速规则，enable 为 true 时此项有效。https://cloud.baidu.com/doc/CDN/s/gjwvyex4o#%E8%AF%B7%E6%B1%82%E4%BD%93
            enable: true
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

