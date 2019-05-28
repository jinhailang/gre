[![Build Status](https://travis-ci.org/jinhailang/gre.svg?branch=master)](https://travis-ci.org/jinhailang/gre)[![codecov](https://codecov.io/gh/jinhailang/gre/branch/master/graph/badge.svg)](https://codecov.io/gh/jinhailang/gre)[![Go Report Card](https://goreportcard.com/badge/github.com/jinhailang/gre)](https://goreportcard.com/report/github.com/jinhailang/gre)

# Introduction

`GRE`(go rule engine) 是使用 Go 语言实现的规则引擎。规则引擎接受输入规则表达式字符串，以及定义数据源，规则表达式解析是基于 Go 抽象语法树（[go/ast](https://godoc.org/go/ast)）实现的，支持大部份的 Go 表达式作为规则表达式使用。表达式语法简洁，强大。基于表达式的实现方式，使 gre 规则引擎具有非常好的扩展性和通用性，并易于理解，使用户更易接受，快速配置正确的规则。

目前，使用 Go 实现的规则引擎开源项目很少，且大多引擎只接受模板化的 JSON 配置，不太灵活，通用性较差，使用起来也比较复杂。gre 期望能提供更简洁，强大的实现方案。我在[规则引擎的实现](https://github.com/jinhailang/blog/issues/49)较详细的阐述了规则引擎的实现及基础原理。

## TDDO

- 直接支持表达式 `when(条件)then(action)`

## Example

```
package main

import (
    "fmt"
    "github.com/jinhailang/gre"
)

func main() {
    dataSource := map[string]interface{}{
	"it":  100,
        "str": "abc",
        "ar":  []int{1, 23, 45},
    }

    rst, err := engine.Run(`(it>ar[2])&&(str=="abc")`, dataSource)
    if err != nil {
    	fmt.Printf("run expr error: %v", err)
    	return
    }

    fmt.Printf("result: %v", rst)
}

// result: true
```

大多数场景下，规则引擎将消息数据（一般是 Json 结构）当作规则主要数据源，代码一般如下：

```
package main

import (
    "fmt"
    "github.com/jinhailang/gre"
    "encoding/json"
)

func main() {
    var dataSource := map[string]interface{}
    var msg := `{"it":100,"str":"abc","ar":[1,23,45]}`

    err := json.Unmarshal([]byte(msg), &dataSource)
    if err != nil {
        fmt.Printf("json error: %v", err)
	return
    }
    
    rst, err := engine.Run(`(it>ar[2])&&(str=="abc")`, dataSource)
    if err != nil {
    	fmt.Printf("run expr error: %v", err)
    	return
    }

    fmt.Printf("result: %v", rst)
}

// result: true
```

**但是，这里有个坑，json.Unmarshal 默认会将 interface 类型变量对应的 json 数字字符串都解析成 float64 类型，而不是 int64**。我在[json.Unmarshal 奇怪的坑](https://github.com/jinhailang/blog/issues/50)作了具体说明以及解决方案。

## Test

```
root@debianvm# go test -cover -bench . -benchtime 3s -benchmem -v
=== RUN   TestEvalBasicLit
--- PASS: TestEvalBasicLit (0.00s)
=== RUN   TestEvalIdent
--- PASS: TestEvalIdent (0.00s)
=== RUN   TestEvalIndexExpr
--- PASS: TestEvalIndexExpr (0.00s)
=== RUN   TestEvalBinaryExpr
--- PASS: TestEvalBinaryExpr (0.00s)
=== RUN   TestEvalUnaryExpr
--- PASS: TestEvalUnaryExpr (0.00s)
=== RUN   TestEvalCallExpr
--- PASS: TestEvalCallExpr (0.00s)
=== RUN   TestRun
--- PASS: TestRun (0.00s)
        engine_test.go:201: expected result. error: interface conversion: interface {} is int64, not string
goos: linux
goarch: amd64
BenchmarkRun-2           1000000              3760 ns/op            1720 B/op         43 allocs/op
PASS
coverage: 71.8% of statements
ok      _/engine 3.874s

```

## Docs

表达式分成三类：

- 一元表达式
- 二元表达式
- 基本表达式（[primary expressions](https://golang.org/ref/spec#Primary_expressions)），即操作数

### 基本表达式

基本表达式是语法上的概念，其实就是数学概念里面的操作数，是运算符作用于的实体。需要注意的是，在语法上，单独的操作数
也被称为表达式。规则引擎支持的基本表达式如下：

- SelectorExp 选择表达式，实例：T.x, T.f()
- ParenExpr 括号内表达式，实例：(!true),(x+y)
- IndexExpr 索引表达式，支持对数组，Map，Slice 索引。例：a[1], mp["x"]
- CallExpr 方法调用表达式，目前支持的调用方法如下：
  - contains 字符串包含，返回 true/false，`contains("abc","a")`
  - matchString 正则匹配，返回 true/false，`matchString("%s+","12dasd")`
  - findAllString 返回所有匹配的项（字符串数组）
  - newSlice 创建字符串 slice，例：`newSlice("a","b","123")`
- BasicLit 文字，包括所有的基本类型，底层结构主要有两属性：类型及名称。实例："123", `"abc"`等
- Ident 标志符，即变量，指在数据源内定义的变量名。若数据源：mp = {"a":"x","b":1}，则 `a`=="x", `b`==1

*关于基本表达式的详细阐述，参考 [go primary expressions](https://golang.org/ref/spec#Primary_expressions)*

### 一元表达式

一元表达式是由操作符与一个操作数组成的表达式。
规则引擎支持的一元操作符如下：

- `!` 表示逻辑否定，操作数必须是 bool 类型，例：!false
- `-` 表示负数，操作数可以是整数或浮点数，例：-0.123，-123

### 二元表达式

二元表达式是由操作符与左，右两个操作数组成的表达式。二元表达式比较常用，二元表达式可以组成（递归）多元表达式。
规则引擎支持的二元操作符如下：

- 基本算术运算符：`+`,`-`,`*`,`/`,`%`
- 比较运算符：`==`,`!=`,`<`,`<=`,`>`,`>=`
- 逻辑运算符：`||`,`&&`

### Other

- 浮点数的运算进行了优化，降低误差，并避免出现类似 `1.223400012-1.2` 计算结果不等于 `0.023400012` 的情况。
