---
Title: mkWeb Template
Template: index.tmpl
# anything else here is up to you and will be passed
# on to the template
---
# Welcome to the mkweb template

```go
package main

import "fmt"

func main() {
    fmt.Println("This is just a basic demo of the features.")
}

```

## MathJax support is also included

$$
\left[ \begin{array}{a} a^l_1 \\ ⋮ \\ a^l_{d_l} \end{array}\right]
= \sigma(
 \left[ \begin{matrix} 
    w^l_{1,1} & ⋯  & w^l_{1,d_{l-1}} \\  
    ⋮ & ⋱  & ⋮  \\ 
    w^l_{d_l,1} & ⋯  & w^l_{d_l,d_{l-1}} \\  
 \end{matrix}\right]  ·
 \left[ \begin{array}{x} a^{l-1}_1 \\ ⋮ \\ ⋮ \\ a^{l-1}_{d_{l-1}} \end{array}\right] + 
 \left[ \begin{array}{b} b^l_1 \\ ⋮ \\ b^l_{d_l} \end{array}\right])
$$

## Where to go from now?

Wherever you want to. Play around a bit with the template, it also supports live-reloading whenever it changes on-disk.
