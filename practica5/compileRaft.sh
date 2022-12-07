CGO_ENABLED=0 go build -o . ./cmd/srvraft/...
mv srvraft raft
docker build . -t localhost:5001/raft:latest
docker push localhost:5001/raft:latest
