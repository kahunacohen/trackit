## Test
To test in an isolated ubantu environment:
1. In the root dev directory do: `go build && go install`
1. `cp $(which trackit)` test/
1. `cp -r ../data .`
1. `cp ../trackit.yaml .`
1. `docker build -f Dockerfile.ubantu -t test-cli .` 
1. `docker run --rm -it test-trackit`
