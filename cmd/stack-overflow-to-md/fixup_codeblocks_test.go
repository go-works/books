package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	s = ` 8. This is the code in the Text Editor:


    using System;
    
    namespace FirstCsharp
    {
        blast
    }

foo
    not code
blast
`
	exp = ` 8. This is the code in the Text Editor:


~~~
using System;

namespace FirstCsharp
{
    blast
}
~~~

foo
    not code
blast
`
)

func TestFixupCodeBlock(t *testing.T) {
	gotBytes := fixupCodeBlocks([]byte(s))
	got := string(gotBytes)
	got = strings.Replace(got, "```", "~~~", -1)
	assert.Equal(t, exp, got)
}
