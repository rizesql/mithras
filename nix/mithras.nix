{
  pkgs,
  go ? pkgs.go_1_26,
  version ? "0.0.0",
}:
pkgs.buildGoModule.override { inherit go; } {
  pname = "mithras";
  inherit version;
  src = pkgs.lib.cleanSource ../.;

  vendorHash = "sha256-KyRDkLdo3vIZyPKsu+HDcoSBSW+K9Tev1UwtfRS1iIE=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  doCheck = false;

  meta = {
    description = "A self-contained authentication provider";
    homepage = "https://github.com/rizesql/mithras";
    license = pkgs.lib.licenses.mit;
    mainProgram = "mithras";
  };
}
