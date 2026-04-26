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
        toolchain = import ../nix/typst.nix { inherit pkgs; };
      in
      {
        devShells.default = toolchain.devShell;
        formatter = pkgs.nixfmt;
      }
    );
}
