{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { self, ... }@inputs:
    inputs.flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import inputs.nixpkgs { inherit system; };

        version =
          if self ? rev && self ? ref then
            let
              tag = builtins.replaceStrings [ "refs/tags/" ] [ "" ] self.ref;
            in
            if builtins.match "v.*" tag != null then tag else builtins.substring 0 7 self.rev
          else
            "dirty";

        toolchain = import ./nix/go.nix { inherit pkgs; };
        go = toolchain.go;

        mithras = import ./nix/mithras.nix { inherit pkgs go version; };
        # docker = import ./nix/docker.nix { inherit pkgs mithras; };
        checks = import ./nix/checks.nix {
          inherit pkgs go;
          src = ./.;
        };

      in
      {
        devShells.default = toolchain.devShell;

        packages = {
          default = mithras;
          inherit mithras; # docker;
        };

        apps.default = {
          type = "app";
          program = "${mithras}/bin/mithras";
          meta = mithras.meta;
        };

        checks = checks;

        formatter = pkgs.nixfmt;
      }
    );
}
