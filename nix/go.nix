{ pkgs }:
{
  devShell = pkgs.mkShell {
    buildInputs = with pkgs; [
      go_1_26
      gopls
      gotools
    ];
    shellHook = ''
      echo "Go version: $(go version)"
      echo "gopls version: $(gopls version)"
    '';
  };
}
