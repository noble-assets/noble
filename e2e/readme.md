# Testing


We utilize the [interchaintest](https://github.com/strangelove-ventures/interchaintest) testing suite.

All tests are located in the [interchaintest folder](../interchaintest/).

## How to Run tests:

### Requirements:

- Docker (running)
- Golang

1. Create local docker image that contains the `noble` binary. 
If you make any code changes, you'll want to re-make the image before running tests. 

`make local-image`

2. Now we can run the tests. There are two ways to run the test.
   
    a. If you are using VS Code you can simply click the `run test` button above each test.
    For this to work, you may need to install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) to VS Code.

    ![vsCode_runTest](../docs/images/vscode_runtests.png)

    b. Or you can run it from the command line:

    ```
    cd interchaintest
    go test -v -run <NAME_OF_TEST>

    # Example
    go test -timeout 10m -v -run TestCCTP_DepForBurnWithCallerOnEth
    ```
