{ pkgs }:
{
  devShell = pkgs.mkShell {
    buildInputs = with pkgs; [
      go_1_26
      gopls
      gotools
    ];

    packages = with pkgs; [
      just
      golangci-lint

      govulncheck
      gosec
    ];

    shellHook = ''
      echo "Go version: $(go version)"
      echo "gopls version: $(gopls version)"
      echo "golangci-lint version: $(golangci-lint version)"
      echo "govulncheck version: $(govulncheck -version)"
      echo "gosec version: $(gosec --version)"
    '';
  };
}
