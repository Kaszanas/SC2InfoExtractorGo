{
  description = "SC2InfoExtractorGo: extracts data from StarCraft II .SC2Replay files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "SC2InfoExtractorGo";
          version = "0.1.0";

          src = ./.;

          # Run `nix build` once and replace this with the hash Nix reports
          # on mismatch.
          vendorHash = pkgs.lib.fakeHash;

          meta = {
            description = "Extracts data from StarCraft II .SC2Replay files into JSON";
            homepage = "https://github.com/Kaszanas/SC2InfoExtractorGo";
            license = pkgs.lib.licenses.gpl3Only;
            mainProgram = "SC2InfoExtractorGo";
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.golangci-lint
          ];
        };
      });
}
