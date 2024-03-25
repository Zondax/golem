# golem

Zondax's opinionated golang lib for projects

## Test

The project contains unit tests you can execute them by running the following command:

`go test -v [TEST FOLDER PATH]`

### Mocking

To generate mocks:

- install: https://github.com/vektra/mockery
- pull the repo with the correct version of the interface you want to mock
- ` mockery --name FullNode --dir $HOME/go/src/github.com/filecoin-project/lotus/api --output ./rosetta/services/mocks --filename fullnode_mock.go --structname FullNodeMock`
- For usage in tests
```
import "github.com/vektra/mockery"

func TestX(t * testing.T){
    mock := MockedInterface{}   
    mock.On("methodName",...args).Return(val)

    // use mock
}