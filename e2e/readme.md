# E2E Testing

We utilize the [interchaintest](https://github.com/strangelove-ventures/interchaintest) testing suite.

All tests are located in the [E2E folder](../e2e/).

## How to Run tests:

### Requirements:

- Docker (running)
- Golang

1. [heighliner](https://github.com/strangelove-ventures/heighliner) will need to be installed.
Heighliner is a tool used to help streamline the creation of the docker image that includes the noble binary
that is being tested.

If it is not already installed, install heighliner:

```bash
git clone https://github.com/strangelove-ventures/heighliner.git
cd heighliner
go install
```

2. Create local docker image that contains the `noble` binary. 
If you make any code changes, you'll want to re-make the image before running tests. 

From top level folder of repo, run:

`make local-image`

3. Run tests.

    From `e2e` folder:

    ```
    go test -v -run <NAME_OF_TEST>

    # Example
    go test -timeout 10m -v -run ^TestFiatTFUpdateOwner$
    ```

    Note: go test uses regex in the test name. Using the `^` and `$` characters help specify single tests. 