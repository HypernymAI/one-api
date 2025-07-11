 2530  ./one-api --port 3000 --log-dir ./logs
 2531  CC="clang --target=arm64-apple-macos11" CXX="clang++ --target=arm64-apple-macos11" GO111MODULE=on CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o one-api-en
 2532  chmod u+x one-api-en
 2533  ./one-api-en --port 3000 --log-dir ./logs
 2534  ls
 2535  cat go.mod
 2536  cat main.go
 2537  ls
 2538  ./one-api-en --port 3000 --log-dir ./logs