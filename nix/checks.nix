{
  pkgs,
  go ? pkgs.go_1_26,
  src,
}:
let
  cleanSrc = pkgs.lib.cleanSource src;

  goEnv = ''
    export HOME=$TMPDIR
    export XDG_CACHE_HOME=$TMPDIR

    export GOCACHE=$TMPDIR/go-cache
    export GOMODCACHE=$TMPDIR/go-mod
    export GOPATH=$TMPDIR/go

    export GOLANGCI_LINT_CACHE=$TMPDIR/golangci-lint
  '';
in
{
  tests = pkgs.buildGoModule.override { inherit go; } {
    pname = "mithras-tests";
    version = "0.0.0";
    src = cleanSrc;
    vendorHash = "sha256-KyRDkLdo3vIZyPKsu+HDcoSBSW+K9Tev1UwtfRS1iIE=";
    doCheck = true;
  };

  lint =
    pkgs.runCommand "mithras-lint"
      {
        buildInputs = [
          pkgs.golangci-lint
          go
        ];
      }
      ''
        ${goEnv}
        cd ${cleanSrc}
        golangci-lint run ./...
        touch $out
      '';

  fmt =
    pkgs.runCommand "mithras-fmt"
      {
        buildInputs = [ go ];
      }
      ''
        ${goEnv}
        if [ -n "$(gofmt -l ${cleanSrc})" ]; then
          echo "Code is not formatted"
          gofmt -l ${cleanSrc}
          exit 1
        fi
        touch $out
      '';
}
