KiCad Database
===============

KCDB ingests github repositories, indexing `.kicad_mod` files so they can be searched and viewed via an easy web interface.

This code powers `https://kcdb.ciphersink.net`.

*Installation*

Ensure you have  Go 1.10 installed.

```shell
git clone <this repo>
cd <this repo>
export GOPATH=`pwd`
go build -o kcdb kcdb.go
./kcdb --listener :80 #stores database in ./kc.db
```

*Manually adding sources*

`./kcdb add-git-source https://github.com/.../...`


*Legal*

```
Copyright (c) 2018 twitchyliquid64

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
