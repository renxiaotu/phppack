# phppack
php pack/unpack for go

与php兼容的数据打包解包工具

# require 依赖

github.com/renxiaotu/phppack

# 使用说明

**struct打包解包：**

```go
package main

import (
	"bytes"
	"fmt"
	"github.com/renxiaotu/phppack"
)
type myType struct {
	Id       uint32 `pack:"N"`
	Name     string `pack:"a10"`
}

func main() {
	mt := &myType{1,  "renxiaotu"}
	b, err := phppack.PackByStruct(mt)//pack 打包
    if err != nil {
    	fmt.Println(err)
    }
    fmt.Println(b)//[0 0 0 1 114 101 110 120 105 97 111 116 117 0]

    nmt := &myType{}
    err = phppack.UnpackByStruct(nmt, b)//unpack 解包
    if err != nil {
    	fmt.Println(err)
    }
    fmt.Println(nmt)//&{1 renxiaotu}
}
```

**format打包解包：**

```go
package main

import (
	"bytes"
	"fmt"
	"github.com/renxiaotu/phppack"
)

func main() {
	b, err := phppack.PackByFormat("Na10",1,"renxiaotu")//pack 打包
    if err != nil {
    	fmt.Println(err)
    }
    fmt.Println(b)//[0 0 0 1 114 101 110 120 105 97 111 116 117 0]

    mt,err := phppack.UnpackByFormat("NId/a10Name", b)//unpack 解包
    if err != nil {
    	fmt.Println(err)
    }
    fmt.Println(mt)//map[Id:1 Name:renxiaotu]
}
```


