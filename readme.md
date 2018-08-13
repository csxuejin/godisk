## 磁盘分区/格式化工具

#### Step 1

- go get github.com/csxuejin/godisk
- cd $GOPATH/src/github.com/csxuejin/godisk
- make build （编译为 linux 环境下可执行二进制文件 godisk）
- 将编译后的 godisk 文件拷贝到服务

#### Step 2

- ./godisk diskinfo   获取磁盘信息，包括磁盘大小（单位为G），磁盘名称等。运行后会在统计目录下生成 `disk.json` 文件
- 编辑 disk.json 文件，对各个字段说明如下
```
{
  "system": "Linux kodoe-2 4.4.0-131-generic #157-Ubuntu SMP Thu Jul 12 15:51:36 UTC 2018 x86_64 x86_64 x86_64 GNU/Linux", // 操作系统信息
  "format_type": "ext4",  // 磁盘格式化类型
  "disks": [
    {
      "name": "/dev/vdb", // 磁盘名称
      "capacity": 10,     // 磁盘容量大小，单位为 G
      "formated": false,  // 是否已经被格式化
      "need_formate": true     // 是否需要进行格式化。true 代表要将该磁盘分为一个区，然后格式化。
    }
  ]
}
```

#### Step 3
- ./godisk partition  根据 disk.json 文件中的内容对磁盘进行分区和格式化

该命令运行后首先会删除指定磁盘的所有分区，然后将该磁盘重新分为一个区，最后将该磁盘格式化。