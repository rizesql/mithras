{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    inputs:
    inputs.flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import inputs.nixpkgs { inherit system; };
        go = import ./nix/go.nix { inherit pkgs; };
        mithras = import ./nix/mithras.nix { inherit pkgs; };

      in
      {
        devShells = {
          default = go.devShell;
        };

        packages = {
          default = mithras;
          inherit mithras;
        };

        apps = {
          default = {
            type = "app";
            program = "${mithras}/bin/mithras";
          };
        };
      }
    );
}
