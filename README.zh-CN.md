# idinfo

`idinfo` 是一个用于识别和分析 UUID、Snowflake、时间戳、网络标识符以及多种其他 ID 格式的命令行工具。

支持自动识别、强制指定格式、时间比较、JSON 和二进制输出，以及可选的 ID 生成。

[English](README.md)

## 功能

- 自动识别 ID 格式，或强制使用指定格式解析。
- 使用 `--everything` 查看所有成功的格式解释。
- 使用 `--compare` 比较不同时间格式解析出的时间。
- 输出信息卡片、简要摘要、JSON 或原始二进制数据。
- 使用 `-` 从标准输入读取一行 ID。
- 为提供生成器的格式生成 ID。

## 安装

需要 Go 1.24 或更高版本。

从当前主分支安装：

```bash
go install github.com/zcyc/idinfo@main
```

从源码构建：

```bash
git clone https://github.com/zcyc/idinfo.git
cd idinfo
go build -o idinfo .
```

## 用法

```text
idinfo [OPTIONS] <ID>
idinfo [OPTIONS] -
idinfo -g <FORMAT>
```

示例：

```bash
# 自动识别
idinfo 550e8400-e29b-41d4-a716-446655440000

# 强制指定格式
idinfo --force uuid 550e8400-e29b-41d4-a716-446655440000
idinfo --force sf-twitter 1777150623882019211

# 输出 JSON
idinfo --output json 507f1f77bcf86cd799439011

# 从标准输入读取一行
echo 550e8400-e29b-41d4-a716-446655440000 | idinfo -
```

## 参数

| 参数 | 说明 |
| --- | --- |
| `-f, --force <FORMAT>` | 按指定格式解析 ID。 |
| `-o, --output <FORMAT>` | 输出格式：`card`、`short`、`json` 或 `binary`。 |
| `-e, --everything` | 显示所有成功的格式解析结果。 |
| `-c, --compare` | 比较不同格式解析出的时间。 |
| `-a, --alphabet <ALPHABET>` | 为 Sqid 和 Nano ID 指定自定义字母表。 |
| `-r, --relative` | 有可用时间戳时显示相对时间。 |
| `--salt <SALT>` | 为 Hashid 指定自定义 salt。 |
| `--epoch <SECONDS>` | 覆盖时间类 ID 使用的 epoch。 |
| `-g, --generate <FORMAT>` | 使用指定格式生成 ID。 |
| `-V, --version` | 显示版本。 |
| `-h, --help` | 显示帮助。 |

`--everything` 会输出专用的信息卡片式多结果内容；`--compare` 用于比较时间类解析结果。

## 支持的格式

以下是 `--force` 接受的 canonical 名称：

```text
uuid shortuuid uuid-int uuid-b64 uuid25
ulid sandflake julid upid comb timeflake flake
scru128 scru64 mongodb ksuid xid cuid1 cuid2 nanoid
tsid sqid hashid youtube stripe datadog nuid typeid
breezeid puid pushid tid threads duns asin snowid
gdocs slack spotify nano64 orderlyid swhid iban commerce vin
bitcoin ethereum
sf-twitter sf-mastodon sf-discord sf-instagram sf-linkedin
sf-sony sf-spaceflake sf-frostflake sf-flakeid sf-simpleflake
mist unix unix-s unix-ms unix-us unix-ns hash ipfs
ipv4 ipv6 mac imei isbn h3
```

支持的 ID 类型包括：

- UUID v1–v8、Nil UUID、Max UUID、NCS UUID 和 Microsoft GUID。
- ShortUUID、UUID Base64、Uuid25 和 UUID 整数表示。
- ULID、Julid、UPID、Sandflake、SCRU128、SCRU64、Timeflake、Flake 和 COMB。
- MongoDB ObjectId、KSUID、Xid、CUID1、CUID2、Nano ID、Sqid、Hashid 和 TypeID。
- TSID、TID、Threads、SnowID、NUID、PUID、PushID、OrderlyID 和 Nano64。
- Twitter、Mastodon、Discord、Instagram、LinkedIn、Sony 等 Snowflake 变体。
- DUNS、ASIN、Google Docs、Slack、Spotify、SWHID、IBAN、ISBN、VIN 和商业条码。
- Bitcoin、Ethereum、IPFS、IPv4、IPv6、MAC、IMEI、H3 和十六进制 Hash。

## 输出

默认输出为信息卡片：

```text
ID Type: UUID (RFC-4122)
Version: 4 (random)
String: 550e8400-e29b-41d4-a716-446655440000
Integer: 113059749145936325402354257176981405696
Size: 128 bits (from hex)
Entropy: 122 bits
```

JSON 输出适合脚本处理：

```bash
idinfo --output json 550e8400-e29b-41d4-a716-446655440000
```

输出简要摘要或原始字节：

```bash
idinfo --output short 550e8400-e29b-41d4-a716-446655440000
idinfo --output binary 550e8400-e29b-41d4-a716-446655440000 | xxd
```

## 常用场景

```bash
# 查看所有成功的格式解析结果
idinfo --everything 1777150623882019211

# 比较时间类解析结果
idinfo --compare 1777150623882019211

# 显示相对时间
idinfo --relative 01JCXSGZMZQQJ2M93WC0T8KT02

# 使用自定义 Hashid salt
idinfo --force hashid --salt 'my secret' 80JTEquWr
```

## 生成 ID

生成是 idinfo 的扩展功能。

支持生成以下 UUID 版本：

```bash
idinfo --generate uuid       # UUID v4
idinfo --generate uuid:v1
idinfo --generate uuid:v3
idinfo --generate uuid:v4
idinfo --generate uuid:v5
idinfo --generate uuid:v6
idinfo --generate uuid:v7
```

其他格式在其实现提供生成器时也可以生成：

```bash
idinfo --generate ulid
idinfo --generate mongodb
```

UUID v8 当前可以解析，但不能通过 `--generate` 生成。

## 开发

```bash
go test ./...
go test -race ./...
go vet ./...
```

## License

MIT License。

## 灵感来源

- [uuinfo](https://github.com/Racum/uuinfo)
