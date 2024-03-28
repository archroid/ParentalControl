# ParentalControl
My miniapp to control my sister when she's playing minecraft🗿

Build for windows X64:
```
env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build  -ldflags="-H windowsgui"
```

NOTE: you need to install ``mingw-w64-gcc`` C compiler!
