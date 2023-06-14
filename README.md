## 中间件 汇总



### Traefik


```shell
docker build -t traefik-plugin:v1 .
```

```shell
docker run -it --rm -p 80:80 -p 8080:8080 -p 8082:8082 -v /d/data/docker/traefik/traefik.yaml:/etc/traefik/traefik.yaml traefik-plugin:v1
```

#### 增加路由

```shell
set traefik-elc-mac/http/routers/micro-elc-orgstruc/rule "Host(`localhost`) && PathPrefix(`/orgstruc/`)"
set traefik-elc-mac/http/routers/micro-elc-orgstruc/service micro-elc-orgstruc
set traefik-elc-mac/http/services/micro-elc-orgstruc/loadBalancer/servers/0/url http://www.baidu.com

# 黑名单
set traefik-elc-mac/http/routers/micro-elc-orgstruc/middlewares/0 my-ipblacklist
set traefik-elc-mac/http/middlewares/my-ipblacklist/plugin/traefik-plugin-ipblacklist/sourcerange 127.0.0.1
set traefik-elc-mac/http/middlewares/my-ipblacklist/plugin/traefik-plugin-ipblacklist/ipstrategy/depth 1
```
