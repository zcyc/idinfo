# idinfo

`idinfo` 是一个用于识别和分析各种 ID 的命令行工具，支持自动识别、强制指定格式、时间比较、JSON/二进制输出，以及可选的 ID 生成。

核心解析命令与 [uuinfo](https://github.com/Racum/uuinfo) 对齐。默认输出信息卡片，也可以输出适合脚本处理的 JSON 或原始二进制数据。

## 安装

需要 Go 1.24 或更高版本。

```bash
go install github.com/zcyc/idinfo@latest
```

也可以从源码构建：

```bash
git clone https://github.com/zcyc/idinfo.git
cd idinfo
go build -o idinfo .
```

## 快速开始

```bash
# 自动识别
idinfo 550e8400-e29b-41d4-a716-446655440000

# 强制按指定格式解析
idinfo -f uuid 550e8400-e29b-41d4-a716-446655440000
idinfo -f sf-twitter 1777150623882019211

# 输出 JSON
idinfo -o json 507f1f77bcf86cd799439011

# 从标准输入读取一行
echo 550e8400-e29b-41d4-a716-446655440000 | idinfo -
```

## 命令格式

```text
idinfo [OPTIONS] <ID>
idinfo [OPTIONS] -
idinfo -g <FORMAT>
```

### 选项

| 选项 | 说明 |
| --- | --- |
| `-f, --force <FORMAT>` | 强制使用指定格式解析 |
| `-o, --output <FORMAT>` | 输出格式：`card`、`short`、`json`、`binary` |
| `-e, --everything` | 显示所有成功的格式解析结果；此模式始终输出卡片 |
| `-c, --compare` | 比较所有时间相关解析结果 |
| `-a, --alphabet <ALPHABET>` | 为 Sqid 或 Nano ID 指定自定义字母表 |
| `-r, --relative` | 在卡片中显示相对时间 |
| `--salt <SALT>` | 为 Hashid 指定自定义 salt |
| `--epoch <SECONDS>` | 为基于时间的 ID 指定 epoch 偏移，单位为秒 |
| `-g, --generate <FORMAT>` | 生成一个 ID；这是 idinfo 的扩展功能 |
| `-V, --version` | 显示版本 |
| `-h, --help` | 显示帮助 |

卡片输出在终端中会自动使用 ANSI 颜色；重定向到文件或管道时通常会自动关闭颜色。

## 支持的格式

### uuinfo 对齐的标准格式

以下名称是强制解析时使用的 canonical 名称：

```text
uuid shortuuid uuid-int uuid-b64 uuid25
ulid sandflake julid upid comb timeflake flake
scru128 scru64 mongodb ksuid xid cuid1 cuid2
nanoid tsid sqid hashid youtube stripe datadog nuid typeid
breezeid puid pushid tid threads duns asin snowid
gdocs slack spotify nano64 orderlyid swhid iban commerce vin
bitcoin ethereum
sf-twitter sf-mastodon sf-discord sf-instagram sf-linkedin
sf-sony sf-spaceflake sf-frostflake sf-flakeid sf-simpleflake
mist unix unix-s unix-ms unix-us unix-ns hash ipfs
ipv4 ipv6 mac imei isbn h3
```

支持的 ID 类型包括：

- UUID 1–8、Nil UUID、Max UUID、NCS UUID 和 Microsoft GUID
- UUID 的 ShortUUID、Base64、Uuid25 和整数表示
- ULID、Julid、UPID、Sandflake、SCRU128、SCRU64、Timeflake、Flake、COMB
- MongoDB ObjectId、KSUID、Xid、CUID1、CUID2、Nano ID、Sqid、Hashid、TypeID
- Twitter、Discord、Instagram、LinkedIn、Sony、Mastodon 等 Snowflake 变体
- TSID、TID、Threads、SnowID、NUID、PUID、PushID、OrderlyID、Nano64
- DUNS、ASIN、Google Docs、Slack、Spotify、SWHID、IBAN、ISBN、VIN 和商业条码
- Bitcoin、Ethereum、IPFS、IPv4、IPv6、MAC、IMEI、H3 和十六进制 Hash

### idinfo 扩展格式

`base58`、`base32`、`snowflake`、`isbn10` 和 `shortpuid` 是 idinfo 的额外解析名称，不属于 uuinfo 的 canonical `--force` 选项；其中 `base58` 和 `base32` 只能通过强制格式使用。

## 输出示例

### Card（默认）

```text
┏━━━━━━━━━━━┯━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ID Type   │ UUID (RFC-4122)                             ┃
┃ Version   │ 4 (random)                                  ┃
┠───────────┼─────────────────────────────────────────────┨
┃ String    │ 550e8400-e29b-41d4-a716-446655440000        ┃
┃ Integer   │ 113059749145936325402354257176981405696     ┃
┠───────────┼─────────────────────────────────────────────┨
┃ Size      │ 128 bits (from hex)                         ┃
┃ Entropy   │ 122 bits                                    ┃
┃ Timestamp │ -                                           ┃
┃ Node 1    │ -                                           ┃
┃ Node 2    │ -                                           ┃
┃ Sequence  │ -                                           ┃
┠───────────┼─────────────────────────────────────────────┨
┃ 550e 8400 │ 0101 0101  0000 1110   1000 0100  0000 0000 ┃
┃ e29b 41d4 │ 1110 0010  1001 1011   0100 0001  1101 0100 ┃
┃ a716 4466 │ 1010 0111  0001 0110   0100 0100  0110 0110 ┃
┃ 5544 0000 │ 0101 0101  0100 0100   0000 0000  0000 0000 ┃
┗━━━━━━━━━━━┷━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### JSON

```json
{
  "id_type": "UUID (RFC-4122)",
  "version": "4 (random)",
  "standard": "550e8400-e29b-41d4-a716-446655440000",
  "integer": 113059749145936325402354257176981405696,
  "uuid_wrap": null,
  "parsed": "from hex",
  "size": 128,
  "entropy": 122,
  "datetime": null,
  "timestamp": null,
  "relative_time": null,
  "sequence": null,
  "node1": null,
  "node2": null,
  "node3": null,
  "hex": "550e8400e29b41d4a716446655440000"
}
```

### Short 和 Binary

```bash
idinfo -o short 550e8400-e29b-41d4-a716-446655440000
idinfo -o binary 550e8400-e29b-41d4-a716-446655440000 | xxd
```

## 常用场景

```bash
# 查看同一个数字可能对应的所有高置信度格式
idinfo -e 1777150623882019211

# 比较不同 Snowflake 变体的时间
idinfo --compare 1777150623882019211

# 显示相对时间
idinfo -r 01JCXSGZMZQQJ2M93WC0T8KT02

# 使用自定义 Hashid salt
idinfo -f hashid --salt 'my secret' 80JTEquWr
```

## 生成 ID

生成是 idinfo 的额外能力，uuinfo 本身不提供 `-g`。UUID 支持以下版本：

```bash
idinfo -g uuid       # UUID v4
idinfo -g uuid:v1
idinfo -g uuid:v3
idinfo -g uuid:v4
idinfo -g uuid:v5
idinfo -g uuid:v6
idinfo -g uuid:v7
```

其他格式只有在实现了生成器时才能使用 `-g`；不支持生成的格式会返回错误。UUID v8 当前可以解析，但不能通过 `-g` 生成。

## 开发

```bash
go test ./...
go test -race ./...
go vet ./...
```

## License

本项目使用 MIT License。

## 致谢

- 设计和格式参考 [uuinfo](https://github.com/Racum/uuinfo)
- 感谢 Go 社区提供的 ID 解析和生成库
