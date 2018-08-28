## 磁盘分区/格式化工具

- go get github.com/csxuejin/godisk

#### 获取磁盘信息

获取到的磁盘信息内容如 disk.json 文件所示。

``` go
import github.com/csxuejin/godisk

....

client := godisk.New()
if data, err := client.GetDiskInfo(); err != nil {
  // handle error
}

```

#### 磁盘分区

``` go

import github.com/csxuejin/godisk

...

client := godisk.New()
if err := client.DiskPartition(result *Result); err != nil {
  // handle error
}

```

该命令运行后首先会删除指定磁盘的所有分区，然后将该磁盘重新分为一个区，最后将该磁盘格式化（默认为 ext4）。
