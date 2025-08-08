# Minimal Go development environment with latest go version and delve debugger.
{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    delve
    goreleaser
  ];

  # Isolate project specific dependencies from the rest of the system.
  shellHook = ''
    export GOPATH=$(pwd)/.go
    export GOBIN=$GOPATH/bin
    export PATH=$GOBIN:$PATH
  '';
}
