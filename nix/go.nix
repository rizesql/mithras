{
  pkgs,
  go ? pkgs.go_1_26,
}:
{
  inherit go;

  devShell = pkgs.mkShell {
    packages = with pkgs; [
      go
      gopls
      gotools
      delve
      golangci-lint

      goreleaser
      hl-log-viewer
      just
    ];

    shellHook = ''
      set -e
      echo "Go:            $(go version | awk '{print $3, $4}')"
      echo "gopls:         $(gopls version | head -1)"
      echo "golangci-lint: $(golangci-lint version --short)"
      echo "just:          $(just --version)"
    '';
  };
}
