# 列表配置文件.list



```go
package main

import (
	"fmt"
	"https-certificate-info/config-list"
	"os"
)

func main() {
	filename := fmt.Sprintf("%s%s", os.Args[0], ".list")
	conf, err := config_list.Parse(`^[a-z0-9]+.+[a-z0-9]+$`, filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(conf)
}
```




