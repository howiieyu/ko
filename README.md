# Ko

Personal Project. A web framework refer to Gin.

## Get Started

```go
package main

import (
	"net/http"
	"github.com/howiieyu/ko"
)

func main() {

	r := ko.Default()
	r.GET("/", func(c *ko.Context) {
		c.String(http.StatusOK, "Hello World\n")
	})

	r.Run(":8080")
}
```
