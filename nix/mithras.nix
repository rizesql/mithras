{ pkgs }:

pkgs.buildGoModule {
  pname = "mithras";
  version = "0.0.0";
  src = ../.;

  subPackages = [ "cmd/mithras" ];

  vendorHash = "sha256-lMELol//HudCZk0BdKdfsbJ1y2r6dWMxfYfnskKdJMo=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=0.0.0"
  ];

  meta = {
    mainProgram = "mithras";
    description = "A self-contained authentication provider";
  };
}
