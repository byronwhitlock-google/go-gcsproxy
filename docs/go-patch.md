## GO Standard Lib Patch

For now, GO-GCSPROXY needs to use a patched go, net/http is updated to handle TE:identity. 

Follow the instruction below to build a patched go command and toolchain:
1. ```git clone git@github.com:golang/go.git```
2. ```git checkout go1.23.0``` go-gcsproxy uses 1.23
3. Make changes to [your-go-repo-root]/src/net/http/transfer.go. Search "eshen" in [here](./go-net-http-patch/transfer.go) as refernce. 
4. Run ```make.bash``` under [your-go-repo-root]/src/. It generates go command under bin/ and toolchain under pkg/
5. Add [your-go-repo-root]/bin/ into PATH env, so the patched go command and toolchain will be used when you launch go-gcsproxy.
6. Go to go-gcsproxy directory and launch the proxy(make uses the patched go). 
