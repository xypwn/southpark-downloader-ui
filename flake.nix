{
  description = "Flake for southpark-downloader-ui";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    utils,
  }:
    utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};

        buildDeps = with pkgs; [
          xorg.libX11
          xorg.libXcursor
          xorg.libXrandr
          xorg.libXinerama
          xorg.libXi
          xorg.libXxf86vm
          libGL
        ];

        runtimeDeps = with pkgs; [
          libGL
          xorg.libX11
        ];
      in {
        packages.default = pkgs.buildGoModule {
          pname = "southpark-downloader-ui";
          version = "latest";

          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-YK0vZpMdxSeKPXY/qRVX4eNrEcW2rpcxySudglwK4oc=";
          subPackages = ["cmd/southpark-downloader-ui"];

          nativeBuildInputs = [pkgs.pkg-config pkgs.makeWrapper];
          buildInputs = buildDeps;
        };
      }
    );
}
