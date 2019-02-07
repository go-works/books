package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	s1 = ` 8. This is the code in the Text Editor:


    using System;
    
    namespace FirstCsharp
    {
        blast
    }

foo
    not code
blast
`
	exp1 = ` 8. This is the code in the Text Editor:


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

	s2 = `    var age = GetAge(dateOfBirth);
    //the above calls the function GetAge passing parameter dateOfBirth.`
	exp2 = `~~~
var age = GetAge(dateOfBirth);
//the above calls the function GetAge passing parameter dateOfBirth.
~~~`

	s3 = `
      var age = GetAge(dateOfBirth);
      //the above calls the function GetAge passing parameter dateOfBirth.
`
	exp3 = `
~~~
var age = GetAge(dateOfBirth);
//the above calls the function GetAge passing parameter dateOfBirth.
~~~
`

	s4 = `
    var age = GetAge(dateOfBirth);

    //the above calls the function GetAge passing parameter dateOfBirth.

t
`
	exp4 = `
~~~
var age = GetAge(dateOfBirth);

//the above calls the function GetAge passing parameter dateOfBirth.
~~~

t
`

	s5 = `
    var age = GetAge(dateOfBirth);
    var age = GetAge(dateOfBirth);


    //the above calls the function GetAge passing parameter dateOfBirth.

bah

`
	exp5 = `
~~~
var age = GetAge(dateOfBirth);
var age = GetAge(dateOfBirth);
~~~


    //the above calls the function GetAge passing parameter dateOfBirth.

bah

`
)

func TestFixupCodeBlock(t *testing.T) {
	f := func(s, exp string) {
		gotBytes := fixupCodeBlocks([]byte(s))
		got := string(gotBytes)
		got = strings.Replace(got, "```", "~~~", -1)
		assert.Equal(t, exp, got)
	}
	f(s1, exp1)
	f(s2, exp2)
	f(s3, exp3)
	f(s4, exp4)
	f(s5, exp5)
}
