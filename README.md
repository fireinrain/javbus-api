# JavBus API <!-- omit in toc -->

一个自我托管的 [JavBus](https://www.javbus.com) API 服务, 但是golang版本

感谢[javbus-api](https://github.com/ovnrain/javbus-api)
## 部署
1. 下载对应的release，并填写好config.toml 运行即可
2. 容器部署,下载docker-compose.yml
```bash
docker compose -f docker-compose.yml up -d
```

## 效果图
![](samples/img.png)
![](samples/img_1.png)
![](samples/img_2.png)

## 配置内容如下  
```toml

############################################
# Application Configuration (TOML Version)
############################################

[server]
# 服务器端口号
DEBUG_LEVEL = "debug"
SERVER_PORT = 3000


[proxy]
# 代理配置
# HTTP代理地址，格式必须以 http://, https://, socks:// 或 socks5:// 开头
HTTP_PROXY = ""

# HTTPS代理地址，格式必须以 http://, https://, socks:// 或 socks5:// 开头
HTTPS_PROXY = ""

# SOCKS5代理地址，格式必须以 socks:// 或 socks5:// 开头
SOCKS5_PROXY = ""


[admin]
# 管理员配置
ADMIN_USERNAME = "admin"

# 管理员密码
ADMIN_PASSWORD = "admin"


[auth]
# 认证配置
# javbus-api 认证token
JAVBUS_JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

# 会话密钥
JAVBUS_SESSION_SECRET = "javbus-api-session-secret"

############################################
# Database Configuration
############################################

[database]
# 数据库类型: "sqlite", "mysql", "postgres"
DB_TYPE = "sqlite"

# DSN，根据不同数据库填写（暂未实现）
# sqlite 示例: "./data/app.db"
# mysql 示例: "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
# postgresql 示例: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
DB_SERVER_PATH = "./data/app.db"

# 最大连接数（MySQL/PostgreSQL有效）
MAX_OPEN_CONNS = 50

# 最大空闲连接数（MySQL/PostgreSQL有效）
MAX_IDLE_CONNS = 10



```


## Api文档示列

## 权限校验

> **Note**
>
> **本项目默认不开启权限校验，即任何人都可以访问。如果项目部署在公网上，建议开启权限校验，以防止被恶意访问**

权限校验目前有两种方式：

### 1. 使用用户名密码

设置以下环境变量：

```env
ADMIN_USERNAME=your_username
ADMIN_PASSWORD=your_password
# 可选，用于加密 session
JAVBUS_SESSION_SECRET=your_session_secret
```

重新部署项目，首页会出现登录按钮，点击进入登录页面，输入用户名密码即可。如果在未登录的情况下访问 API，会跳转到登录页面。这种方式适合在浏览器中访问，如果是在 App 中访问，建议使用第二种方式

### 2. 使用 Token

这种方式需要在请求头中添加 `j-auth-token` 字段，值为 `your_token`

设置以下环境变量：

```env
JAVBUS_AUTH_TOKEN=your_token
```

重新部署项目，即可使用 Token 访问 API

下面是几个例子：

使用 `curl`

```shell
$ curl -H "Authorization: Bearer your_config_jwt_token" http://localhost:8922/api/stars/okq
```

使用 `Wget`

```shell
$ wget --header="Authorization: Bearer your_config_jwt_token" http://localhost:8922/api/stars/okq
```

使用 [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)

```http
GET http://localhost:8922/api/stars/okq HTTP/1.1
Authorization: Bearer your_config_jwt_token
```


> **Note**
>
> **只设置 `JAVBUS_JWT_TOKEN` 环境变量是不安全的，用户依然可以通过不加 `Authorization` 请求头，或者在浏览器中直接访问 API。因此，应该同时设置 `ADMIN_USERNAME` 和 `ADMIN_PASSWORD` 环境变量，以达到双重校验的目的**

## API 文档

### /api/movies

获取影片列表

#### method

GET

#### 参数

| 参数        | 是否必须 | 可选值                                                                       | 默认值   | 说明                                                                                                                                                              |
| ----------- | -------- | ---------------------------------------------------------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| page        | 否       |                                                                              | `1`      | 页码                                                                                                                                                              |
| magnet      | 否       | `exist`<br />`all`                                                           | `exist`  | `exist`: 只返回有磁力链接的影片<br />`all`: 返回全部影片                                                                                                          |
| filterType  | 否       | `star`<br />`genre`<br />`director`<br />`studio`<br />`label`<br />`series` |          | 筛选类型，必须与 `filterValue` 一起使用<br />`star`: 演员<br />`genre`: 类别<br />`director`: 导演<br />`studio`: 制作商<br />`label`: 发行商<br />`series`: 系列 |
| filterValue | 否       |                                                                              |          | 筛选值，必须与 `filterType` 一起使用                                                                                                                              |
| type        | 否       | `normal`<br />`uncensored`                                                   | `normal` | `normal`: 有码影片<br />`uncensored`: 无码影片                                                                                                                    |

#### 请求举例

    /api/movies

返回有磁力链接的第一页影片

    /api/movies?filterType=star&filterValue=rsv&magnet=all

返回演员 ID 为 `rsv` 的影片的第一页，包含有磁力链接和无磁力链接的影片

    /api/movies?page=2&filterType=genre&filterValue=4

返回类别 ID 为 `4` 的影片的第二页，只返回有磁力链接的影片

    /api/movies?type=uncensored

返回无码影片的第一页，只返回有磁力链接的影片

#### 返回举例

<details>
<summary>点击展开</summary>

```jsonc
{
  // 影片列表
  "movies": [
    {
      "date": "2023-04-28",
      "id": "YUJ-003",
      "img": "https://www.javbus.com/pics/thumb/9n0d.jpg",
      "title": "夫には言えない三日間。 セックスレスで欲求不満な私は甥っ子に中出しさせています。 岬ななみ",
      "tags": ["高清", "字幕", "3天前新種"]
    }
    // ...
  ],
  // 分页信息
  "pagination": {
    "currentPage": 1,
    "hasNextPage": true,
    "nextPage": 2,
    "pages": [1, 2, 3]
  },
  // 筛选信息，注意：只有在请求参数包含 filterType 和 filterValue 时才会返回
  "filter": {
    "name": "岬ななみ",
    "type": "star",
    "value": "rsv"
  }
}
```

</details>

### /api/movies/search

搜索影片

#### method

GET

#### 参数

| 参数    | 是否必须 | 可选值                     | 默认值   | 说明                                                     |
| ------- | -------- | -------------------------- | -------- | -------------------------------------------------------- |
| keyword | 是       |                            |          | 搜索关键字                                               |
| page    | 否       |                            | `1`      | 页码                                                     |
| magnet  | 否       | `exist`<br />`all`         | `exist`  | `exist`: 只返回有磁力链接的影片<br />`all`: 返回全部影片 |
| type    | 否       | `normal`<br />`uncensored` | `normal` | `normal`: 有码影片<br />`uncensored`: 无码影片           |

#### 请求举例

    /api/movies/search?keyword=三上

搜索关键词为 `三上` 的影片的第一页，只返回有磁力链接的影片

    /api/movies/search?keyword=三上&magnet=all

搜索关键词为 `三上` 的影片的第一页，包含有磁力链接和无磁力链接的影片

#### 返回举例

<details>
<summary>点击展开</summary>

```jsonc
{
  // 影片列表
  "movies": [
    {
      "date": "2020-08-15",
      "id": "SSNI-845",
      "img": "https://www.javbus.com/pics/thumb/7t44.jpg",
      "title": "彼女の姉は美人で巨乳しかもドS！大胆M性感プレイでなす術もなくヌキまくられるドMな僕。 三上悠亜",
      "tags": ["高清", "字幕"]
    }
    // ...
  ],
  // 分页信息
  "pagination": {
    "currentPage": 2,
    "hasNextPage": true,
    "nextPage": 3,
    "pages": [1, 2, 3, 4, 5]
  },
  "keyword": "三上"
}
```

</details>

### /api/movies/{movieId}

获取影片详情

#### method

GET

#### 请求举例

    /api/movies/SSIS-406

返回番号为 `SSIS-406` 的影片详情

#### 返回举例

<details>
<summary>点击展开</summary>

```jsonc
{
  "id": "SSIS-406",
  "title": "SSIS-406 才色兼備な女上司が思う存分に羽目を外し僕を連れ回す【週末限定】裏顔デート 葵つかさ",
  "img": "https://www.javbus.com/pics/cover/8xnc_b.jpg",
  // 封面大图尺寸
  "imageSize": {
    "width": 800,
    "height": 538
  },
  "date": "2022-05-20",
  // 影片时长
  "videoLength": 120,
  "director": {
    "id": "hh",
    "name": "五右衛門"
  },
  "producer": {
    "id": "7q",
    "name": "エスワン ナンバーワンスタイル"
  },
  "publisher": {
    "id": "9x",
    "name": "S1 NO.1 STYLE"
  },
  "series": {
    "id": "xx",
    "name": "xx"
  },
  "genres": [
    {
      "id": "e",
      "name": "巨乳"
    }
    // ...
  ],
  // 演员信息，一部影片可能包含多个演员
  "stars": [
    {
      "id": "2xi",
      "name": "葵つかさ"
    }
  ],
  // 影片预览图
  "samples": [
    {
      "alt": "SSIS-406 才色兼備な女上司が思う存分に羽目を外し僕を連れ回す【週末限定】裏顔デート 葵つかさ - 樣品圖像 - 1",
      "id": "8xnc_1",
      // 大图
      "src": "https://pics.dmm.co.jp/digital/video/ssis00406/ssis00406jp-1.jpg",
      // 缩略图
      "thumbnail": "https://www.javbus.com/pics/sample/8xnc_1.jpg"
    }
    // ...
  ],
  // 同类影片
  "similarMovies": [
    {
      "id": "SNIS-477",
      "title": "クレーム処理会社の女社長 土下座とカラダで解決します 夢乃あいか",
      "img": "https://www.javbus.com/pics/thumb/4wml.jpg"
    }
    // ...
  ],
  "gid": "50217160940",
  "uc": "0"
}
```

</details>

### /api/magnets/{movieId}

获取影片磁力链接

#### method

GET

#### 参数

| 参数      | 是否必须 | 可选值             | 默认值 | 说明                                            |
| --------- | -------- | ------------------ | ------ | ----------------------------------------------- |
| gid       | 是       |                    |        | 从影片详情获取到的 `gid`                        |
| uc        | 是       |                    |        | 从影片详情获取到的 `uc`                         |
| sortBy    | 否       | `date`<br />`size` | `size` | 按照日期或大小排序，必须与 `sortOrder` 一起使用 |
| sortOrder | 否       | `asc`<br />`desc`  | `desc` | 升序或降序，必须与 `sortBy` 一起使用            |

#### 请求举例

    /api/magnets/SSNI-730?gid=42785257471&uc=0

返回番号为 `SSNI-730` 的影片的磁力链接

    /api/magnets/SSNI-730?gid=42785257471&uc=0&sortBy=size&sortOrder=asc

返回番号为 `SSNI-730` 的影片的磁力链接，并按照大小升序排序

    /api/magnets/SSNI-730?gid=42785257471&uc=0&sortBy=date&sortOrder=desc

返回番号为 `SSNI-730` 的影片的磁力链接，并按照日期降序排序

#### 返回举例

<details>
<summary>点击展开</summary>

```jsonc
[
  {
    "id": "17508BF5C17CBDF7C77E12DAAD1BDAB325116585",
    "link": "magnet:?xt=urn:btih:17508BF5C17CBDF7C77E12DAAD1BDAB325116585&dn=SSNI-730-C",
    // 是否高清
    "isHD": true,
    "title": "SSNI-730-C",
    "size": "6.57GB",
    // bytes
    "numberSize": 7054483783,
    "shareDate": "2021-03-14",
    // 是否包含字幕
    "hasSubtitle": true
  }
  // ...
]
```

</details>

### /api/stars/{starId}

获取演员详情

#### method

GET

#### 参数

| 参数 | 是否必须 | 可选值                     | 默认值   | 说明                                                           |
| ---- | -------- | -------------------------- | -------- | -------------------------------------------------------------- |
| type | 否       | `normal`<br />`uncensored` | `normal` | `normal`: 有码影片演员详情<br />`uncensored`: 无码影片演员详情 |

#### 请求举例

    /api/stars/2xi

返回演员 `葵つかさ` 的详情

    /api/stars/2jd?type=uncensored

返回演员 `波多野結衣` 的详情

#### 返回举例

<details>
<summary>点击展开</summary>

```jsonc
{
  "avatar": "https://www.javbus.com/pics/actress/2xi_a.jpg",
  "id": "2xi",
  "name": "葵つかさ",
  "birthday": "1990-08-14",
  "age": "32",
  "height": "163cm",
  "bust": "88cm",
  "waistline": "58cm",
  "hipline": "86cm",
  "birthplace": "大阪府",
  "hobby": "ジョギング、ジャズ鑑賞、アルトサックス、ピアノ、一輪車"
}
```

</details>



