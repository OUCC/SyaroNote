Sample Page
====

Title
----
### h3

#### h4

##### h5

###### h6

Plain text
----
Lorem ipsum dolor sit amet, consectetur adipiscing elit. 

Aliquam mi orci, porta vitae nisl sit amet, imperdiet fringilla dolor.

Rich text
----
Lorem *ipsum* **dolor** ***sit*** ~~amet~~

### Link
[google](http://google.co.jp)

### List
* item 1
* item 2
    * item 2-1
        * item 2-1-1
    * item 2-2
* item 3

1. item 1
2. item 2
    1. item 2-1
    2. item 2-2

### Quote
> quote

> > nest

> end

### Table

| Left align | Right align | Center align |
|:-----------|------------:|:------------:|
| This       |        This |     This     |
| column     |      column |    column    |
| will       |        will |     will     |
| be         |          be |      be      |
| left       |       right |    center    |
| aligned    |     aligned |   aligned    |

### Horizontal line

---

Code
----
To use code highlighter, run `syaro --highlight`

```HTML
<!DOCTYPE html>
<head>
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<title></title>
```

```css
body { display: none; }
```

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, golang")
}
```

`<code></code>`

WikiLink
----
* [[Home]]
* [[page not available]]

Math
----
To use [MathJax](http://www.mathjax.org/), run `syaro --mathjax`.

inline math $\mathrm{e}^{i\theta}=\cos\theta+i\sin\theta$

$$ S=\sum^\infty_{n=1}s_n $$
