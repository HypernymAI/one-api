#CC="clang --target=arm64-apple-macos11" CXX="clang++ --target=arm64-apple-macos11" GO111MODULE=on CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o one-api-en
#chmod u+x one-api-en
./one-api-en --port 3000 --log-dir ./logs