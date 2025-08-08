# Minimal Go development environment with latest go version and delve debugger.
{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    delve
    goreleaser
  ];

  shellHook = ''
    # Isolate project specific dependencies from the rest of the system.
    export GOPATH=$(pwd)/.go
    export GOBIN=$GOPATH/bin
    export PATH=$GOBIN:$PATH

    ### Go Releaser Logic ###

    export GITHUB_TOKEN=$GITHUB_TOKEN

    release() {
      if [ -z "$GITHUB_TOKEN" ]; then
        echo "Error: GITHUB_TOKEN is not set."
        echo "Please source your .env file or export it."
        return 1
      fi
      
      goreleaser release --clean
    }

    ### Test ###

    test() {
      go test ./...
    }
  '';
}
