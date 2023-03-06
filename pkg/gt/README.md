# gt

## 功能介绍
    http 请求工具，支持 Restfull 风格的请求，同时支持将respons数据序列化到结构体

## example
用例见client_test.go

## 当前支持返还值绑定类型
- JSON
- YAML
- BODY (string, []byte) 

## header 添加
- 单个key SetHeader(key string, values ...string)
- 多个key AddHeader(header http.Header)