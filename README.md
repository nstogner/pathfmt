# pathfmt

Parse variables from paths like `/ids/{id}` in Go.

## Structs

```go
import (
	"fmt"
	"log"

	"github.com/nstogner/pathfmt"
)

func main() {
	type UserIdentifier struct {
		OrgNum int    `path:"org_num"`
		ID     string `path:"id"`
	}

	f := pathfmt.New("/organizations/{org_num}/users/{id}")

	var u UserIdentifier
	if err := f.ToStruct("/organizations/123/users/nick", &u); err != nil {
		log.Fatal(err)
	}

	fmt.Println(u.OrgNum, u.ID)
	// Output: 123 nick
}
```

## Maps

```go
import (
	"fmt"
	"log"

	"github.com/nstogner/pathfmt"
)

func main() {
	f := pathfmt.New("/api/v1/users/{id}")

	m, err := f.ToMap("/api/v1/users/123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(m)
	// Output: map[id:123]
}
```
